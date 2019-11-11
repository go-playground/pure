package middleware

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	httpext "github.com/go-playground/pkg/v4/net/http"

	. "github.com/go-playground/assert/v2"
	"github.com/go-playground/pure/v5"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//

func TestGzip(t *testing.T) {

	p := pure.New()
	p.Use(Gzip)
	p.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("test"))
	})
	p.Get("/empty", func(w http.ResponseWriter, r *http.Request) {
	})

	server := httptest.NewServer(p.Serve())
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)

	client := &http.Client{}

	resp, err := client.Do(req)
	Equal(t, err, nil)
	Equal(t, resp.StatusCode, http.StatusOK)

	b, err := ioutil.ReadAll(resp.Body)
	Equal(t, err, nil)
	Equal(t, string(b), "test")

	req, _ = http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	req.Header.Set(httpext.AcceptEncoding, "gzip")

	resp, err = client.Do(req)
	Equal(t, err, nil)
	Equal(t, resp.StatusCode, http.StatusOK)
	Equal(t, resp.Header.Get(httpext.ContentEncoding), httpext.Gzip)
	Equal(t, resp.Header.Get(httpext.ContentType), httpext.TextPlain)

	r, err := gzip.NewReader(resp.Body)
	Equal(t, err, nil)
	defer r.Close()

	b, err = ioutil.ReadAll(r)
	Equal(t, err, nil)
	Equal(t, string(b), "test")

	req, _ = http.NewRequest(http.MethodGet, server.URL+"/empty", nil)

	resp, err = client.Do(req)
	Equal(t, err, nil)
	Equal(t, resp.StatusCode, http.StatusOK)
}

func TestGzipLevel(t *testing.T) {

	// bad gzip level
	PanicMatches(t, func() { GzipLevel(999) }, "gzip: invalid compression level: 999")

	p := pure.New()
	p.Use(GzipLevel(flate.BestCompression))
	p.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("test"))
	})
	p.Get("/empty", func(w http.ResponseWriter, r *http.Request) {
	})

	server := httptest.NewServer(p.Serve())
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)

	client := &http.Client{}

	resp, err := client.Do(req)
	Equal(t, err, nil)
	Equal(t, resp.StatusCode, http.StatusOK)

	b, err := ioutil.ReadAll(resp.Body)
	Equal(t, err, nil)
	Equal(t, string(b), "test")

	req, _ = http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	req.Header.Set(httpext.AcceptEncoding, "gzip")

	resp, err = client.Do(req)
	Equal(t, err, nil)
	Equal(t, resp.StatusCode, http.StatusOK)
	Equal(t, resp.Header.Get(httpext.ContentEncoding), httpext.Gzip)
	Equal(t, resp.Header.Get(httpext.ContentType), httpext.TextPlain)

	r, err := gzip.NewReader(resp.Body)
	Equal(t, err, nil)
	defer r.Close()

	b, err = ioutil.ReadAll(r)
	Equal(t, err, nil)
	Equal(t, string(b), "test")

	req, _ = http.NewRequest(http.MethodGet, server.URL+"/empty", nil)

	resp, err = client.Do(req)
	Equal(t, err, nil)
	Equal(t, resp.StatusCode, http.StatusOK)
}

func TestGzipFlush(t *testing.T) {

	rec := httptest.NewRecorder()
	buff := new(bytes.Buffer)

	w := gzip.NewWriter(buff)
	gw := gzipWriter{Writer: w, ResponseWriter: rec}

	Equal(t, buff.Len(), 0)

	err := gw.Flush()
	Equal(t, err, nil)

	n1 := buff.Len()
	NotEqual(t, n1, 0)

	_, err = gw.Write([]byte("x"))
	Equal(t, err, nil)

	n2 := buff.Len()
	Equal(t, n1, n2)

	err = gw.Flush()
	Equal(t, err, nil)
	NotEqual(t, n2, buff.Len())
}

func TestGzipHijack(t *testing.T) {

	rec := newCloseNotifyingRecorder()
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	gw := gzipWriter{Writer: w, ResponseWriter: rec}

	_, bufrw, err := gw.Hijack()
	Equal(t, err, nil)

	_, _ = bufrw.WriteString("test")
}

type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func newCloseNotifyingRecorder() *closeNotifyingRecorder {
	return &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

func (c *closeNotifyingRecorder) Close() {
	c.closed <- true
}

func (c *closeNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}

func (c *closeNotifyingRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {

	reader := bufio.NewReader(c.Body)
	writer := bufio.NewWriter(c.Body)
	return nil, bufio.NewReadWriter(reader, writer), nil
}
