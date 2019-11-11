package middleware

import (
	"bufio"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"

	httpext "github.com/go-playground/pkg/v4/net/http"

	"github.com/go-playground/pure/v5"
)

type gzipWriter struct {
	io.Writer
	http.ResponseWriter
	sniffComplete bool
}

func (w *gzipWriter) Write(b []byte) (int, error) {

	if !w.sniffComplete {
		if w.Header().Get(httpext.ContentType) == "" {
			w.Header().Set(httpext.ContentType, http.DetectContentType(b))
		}
		w.sniffComplete = true
	}

	return w.Writer.Write(b)
}

func (w *gzipWriter) Flush() error {
	return w.Writer.(*gzip.Writer).Flush()
}

func (w *gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

var gzipPool = sync.Pool{
	New: func() interface{} {
		return &gzipWriter{Writer: gzip.NewWriter(ioutil.Discard)}
	},
}

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add(httpext.Vary, httpext.AcceptEncoding)

		if strings.Contains(r.Header.Get(httpext.AcceptEncoding), httpext.Gzip) {

			gz := gzipPool.Get().(*gzipWriter)
			gz.sniffComplete = false
			gzr := gz.Writer.(*gzip.Writer)
			gzr.Reset(w)
			gz.ResponseWriter = w

			w.Header().Set(httpext.ContentEncoding, httpext.Gzip)

			w = gz
			defer func() {

				if !gz.sniffComplete {
					// We have to reset response to it's pristine state when
					// nothing is written to body.
					w.Header().Del(httpext.ContentEncoding)
					gzr.Reset(ioutil.Discard)
				}

				gzr.Close()
				gzipPool.Put(gz)
			}()
		}

		next(w, r)
	}
}

// GzipLevel returns a middleware which compresses HTTP response using gzip compression
// scheme using the level specified
func GzipLevel(level int) pure.Middleware {

	// test gzip level, then don't have to each time one is created
	// in the pool

	if _, err := gzip.NewWriterLevel(ioutil.Discard, level); err != nil {
		panic(err)
	}

	var gzipPool = sync.Pool{
		New: func() interface{} {
			z, _ := gzip.NewWriterLevel(ioutil.Discard, level)

			return &gzipWriter{Writer: z}
		},
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			w.Header().Add(httpext.Vary, httpext.AcceptEncoding)

			if strings.Contains(r.Header.Get(httpext.AcceptEncoding), httpext.Gzip) {

				gz := gzipPool.Get().(*gzipWriter)
				gz.sniffComplete = false
				gzr := gz.Writer.(*gzip.Writer)
				gzr.Reset(w)
				gz.ResponseWriter = w

				w.Header().Set(httpext.ContentEncoding, httpext.Gzip)

				w = gz
				defer func() {

					if !gz.sniffComplete {
						// We have to reset response to it's pristine state when
						// nothing is written to body.
						w.Header().Del(httpext.ContentEncoding)
						gzr.Reset(ioutil.Discard)
					}

					gzr.Close()
					gzipPool.Put(gz)
				}()
			}

			next(w, r)
		}
	}
}
