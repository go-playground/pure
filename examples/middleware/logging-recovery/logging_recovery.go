package middleware

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/go-playground/ansi"
	"github.com/go-playground/pure"
)

const (
	status500 = ansi.Underline + ansi.Blink + ansi.Red
	status400 = ansi.Red
	status300 = ansi.Yellow
	status    = ansi.Green
)

type logWriter struct {
	http.ResponseWriter
	status    int
	size      int64
	committed bool
}

// WriteHeader writes HTTP status code.
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
// Thus explicit calls to WriteHeader are mainly used to
// send error codes.
func (lw *logWriter) WriteHeader(status int) {

	if lw.committed {
		log.Println("response already committed")
		return
	}

	lw.status = status
	lw.ResponseWriter.WriteHeader(status)
	lw.committed = true
}

// Write writes the data to the connection as part of an HTTP reply.
// If WriteHeader has not yet been called, Write calls WriteHeader(http.StatusOK)
// before writing the data.  If the Header does not contain a
// Content-Type line, Write adds a Content-Type set to the result of passing
// the initial 512 bytes of written data to DetectContentType.
func (lw *logWriter) Write(b []byte) (int, error) {

	lw.size += int64(len(b))

	return lw.ResponseWriter.Write(b)
}

// Status returns the current response's http status code.
func (lw *logWriter) Status() int {
	return lw.status
}

// Size returns the number of bytes written in the response thus far
func (lw *logWriter) Size() int64 {
	return lw.size
}

// Hijack hijacks the current http connection
func (lw *logWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return lw.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify ...
func (lw *logWriter) CloseNotify() <-chan bool {
	return lw.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

var lrpool = sync.Pool{
	New: func() interface{} {
		return new(logWriter)
	},
}

// LoggingAndRecovery handle HTTP request logging + recovery
func LoggingAndRecovery(color bool) pure.Middleware {

	return func(next http.HandlerFunc) http.HandlerFunc {

		if color {

			return func(w http.ResponseWriter, r *http.Request) {

				t1 := time.Now()

				lw := lrpool.Get().(*logWriter)
				lw.status = 200
				lw.size = 0
				lw.committed = false
				lw.ResponseWriter = w

				defer func() {
					if err := recover(); err != nil {
						trace := make([]byte, 1<<16)
						n := runtime.Stack(trace, true)
						log.Printf(" %srecovering from panic: %+v\nStack Trace:\n %s%s", ansi.Red, err, trace[:n], ansi.Reset)
						HandlePanic(lw, r, trace[:n])

						lrpool.Put(lw)
						return
					}

					lrpool.Put(lw)
				}()

				next(lw, r)

				// var color string
				color := status

				code := lw.Status()

				switch {
				case code >= http.StatusInternalServerError:
					color = status500
				case code >= http.StatusBadRequest:
					color = status400
				case code >= http.StatusMultipleChoices:
					color = status300
				default:
					color = status
				}

				log.Printf("%s %d %s[%s%s%s] %q %v %d\n", color, code, ansi.Reset, color, r.Method, ansi.Reset, r.URL, time.Since(t1), lw.Size())
			}
		}

		return func(w http.ResponseWriter, r *http.Request) {

			t1 := time.Now()

			lw := lrpool.Get().(*logWriter)
			lw.status = 200
			lw.size = 0
			lw.committed = false
			lw.ResponseWriter = w

			defer func() {
				if err := recover(); err != nil {
					trace := make([]byte, 1<<16)
					n := runtime.Stack(trace, true)
					log.Printf(" %srecovering from panic: %+v\nStack Trace:\n %s%s", ansi.Red, err, trace[:n], ansi.Reset)
					HandlePanic(lw, r, trace[:n])
				}

				lrpool.Put(lw)
			}()

			next(lw, r)

			log.Printf("%d [%s] %q %v %d\n", lw.Status(), r.Method, r.URL, time.Since(t1), lw.Size())
		}

	}
}

// HandlePanic handles graceful panic by redirecting to friendly error page or rendering a friendly error page.
// trace passed just in case you want rendered to developer when not running in production
func HandlePanic(w http.ResponseWriter, r *http.Request, trace []byte) {

	// redirect to or directly render friendly error page
}
