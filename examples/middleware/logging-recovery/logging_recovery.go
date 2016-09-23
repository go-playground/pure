package middleware

import (
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// ANSIEscSeq is a predefined ANSI escape sequence
type ANSIEscSeq string

// ANSI escape sequences
// NOTE: in an standard xterm terminal the light colors will appear BOLD instead of the light variant
const (
	Black        ANSIEscSeq = "\x1b[30m"
	DarkGray                = "\x1b[30;1m"
	Blue                    = "\x1b[34m"
	LightBlue               = "\x1b[34;1m"
	Green                   = "\x1b[32m"
	LightGreen              = "\x1b[32;1m"
	Cyan                    = "\x1b[36m"
	LightCyan               = "\x1b[36;1m"
	Red                     = "\x1b[31m"
	LightRed                = "\x1b[31;1m"
	Magenta                 = "\x1b[35m"
	LightMagenta            = "\x1b[35;1m"
	Brown                   = "\x1b[33m"
	Yellow                  = "\x1b[33;1m"
	LightGray               = "\x1b[37m"
	White                   = "\x1b[37;1m"
	Underscore              = "\x1b[4m"
	Blink                   = "\x1b[5m"
	Inverse                 = "\x1b[7m"
	Reset                   = "\x1b[0m"
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

var lrpool = sync.Pool{
	New: func() interface{} {
		return new(logWriter)
	},
}

// LoggingAndRecovery handle HTTP request logging + recovery
func LoggingAndRecovery(next http.HandlerFunc) http.HandlerFunc {

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
				log.Printf(" %srecovering from panic: %+v\nStack Trace:\n %s%s", Red, err, trace[:n], Reset)
				HandlePanic(lw, r, trace[:n])

				lrpool.Put(lw)
				return
			}

			lrpool.Put(lw)
		}()

		next(lw, r)

		var color string

		code := lw.Status()

		switch {
		case code >= http.StatusInternalServerError:
			color = Underscore + Blink + Red
		case code >= http.StatusBadRequest:
			color = Red
		case code >= http.StatusMultipleChoices:
			color = Yellow
		default:
			color = Green
		}

		log.Printf("%s %d %s[%s%s%s] %q %v %d\n", color, code, Reset, color, r.Method, Reset, r.URL, time.Since(t1), lw.Size())
	}
}

// HandlePanic handles graceful panic by redirecting to friendly error page or rendering a friendly error page.
// trace passed just in case you want rendered to developer when not running in production
func HandlePanic(w http.ResponseWriter, r *http.Request, trace []byte) {

	// redirect to or directly render friendly error page
}
