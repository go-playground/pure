package pure

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "gopkg.in/go-playground/assert.v1"
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

func TestUseAndGroup(t *testing.T) {

	fn := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(r.Method)); err != nil {
			panic(err)
		}
	}

	var log string

	logger := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log = r.URL.Path
			next(w, r)
		}
	}

	p := New()
	p.Use(logger)
	p.Get("/", fn)

	code, body := request(http.MethodGet, "/", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, http.MethodGet)
	Equal(t, log, "/")

	g := p.Group("/users")
	g.Get("/", fn)
	g.Get("/list/", fn)

	code, body = request(http.MethodGet, "/users/", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, http.MethodGet)
	Equal(t, log, "/users/")

	code, body = request(http.MethodGet, "/users/list/", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, http.MethodGet)
	Equal(t, log, "/users/list/")

	logger2 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log = r.URL.Path + "2"
			next(w, r)
		}
	}

	sh := p.GroupWithMore("/superheros", logger2)
	sh.Get("/", fn)
	sh.Get("/list/", fn)

	code, body = request(http.MethodGet, "/superheros/", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, http.MethodGet)
	Equal(t, log, "/superheros/2")

	code, body = request(http.MethodGet, "/superheros/list/", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, http.MethodGet)
	Equal(t, log, "/superheros/list/2")

	sc := sh.Group("/children")
	sc.Get("/", fn)
	sc.Get("/list/", fn)

	code, body = request(http.MethodGet, "/superheros/children/", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, http.MethodGet)
	Equal(t, log, "/superheros/children/2")

	code, body = request(http.MethodGet, "/superheros/children/list/", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, http.MethodGet)
	Equal(t, log, "/superheros/children/list/2")

	log = ""

	g2 := p.GroupWithNone("/admins")
	g2.Get("/", fn)
	g2.Get("/list/", fn)

	code, body = request(http.MethodGet, "/admins/", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, http.MethodGet)
	Equal(t, log, "")

	code, body = request(http.MethodGet, "/admins/list/", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, http.MethodGet)
	Equal(t, log, "")
}

func TestMatch(t *testing.T) {

	p := New()
	p.Match([]string{http.MethodConnect, http.MethodDelete, http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPatch, http.MethodPost, http.MethodPut, http.MethodTrace}, "/test", defaultHandler)

	hf := p.Serve()

	tests := []struct {
		method string
	}{
		{
			method: http.MethodConnect,
		},
		{
			method: http.MethodDelete,
		},
		{
			method: http.MethodGet,
		},
		{
			method: http.MethodHead,
		},
		{
			method: http.MethodOptions,
		},
		{
			method: http.MethodPatch,
		},
		{
			method: http.MethodPost,
		},
		{
			method: http.MethodPut,
		},
		{
			method: http.MethodTrace,
		},
	}

	for _, tt := range tests {
		req, err := http.NewRequest(tt.method, "/test", nil)
		if err != nil {
			t.Errorf("Expected 'nil' Got '%s'", err)
		}

		res := httptest.NewRecorder()

		hf.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Errorf("Expected '%d' Got '%d'", http.StatusOK, res.Code)
		}

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Expected 'nil' Got '%s'", err)
		}

		s := string(b)

		if s != tt.method {
			t.Errorf("Expected '%s' Got '%s'", tt.method, s)
		}
	}
}

func TestGrouplogic(t *testing.T) {

	var aa, bb, cc, tl int

	aM := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			aa++
			next(w, r)
		}
	}

	bM := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			bb++
			next(w, r)
		}
	}

	cM := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cc++
			next(w, r)
		}
	}

	p := New()
	p.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			tl++
			next(w, r)
		}
	})

	a := p.GroupWithMore("/a", aM)
	a.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("a-ok"))
	})

	b := a.GroupWithMore("/b", bM)
	b.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("b-ok"))
	})

	c := b.GroupWithMore("/c", cM)
	c.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("c-ok"))
	})

	code, body := request(http.MethodGet, "/a/test", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "a-ok")
	Equal(t, tl, 1)
	Equal(t, aa, 1)

	code, body = request(http.MethodGet, "/a/b/test", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "b-ok")
	Equal(t, tl, 2)
	Equal(t, aa, 2)
	Equal(t, bb, 1)

	code, body = request(http.MethodGet, "/a/b/c/test", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "c-ok")
	Equal(t, tl, 3)
	Equal(t, aa, 3)
	Equal(t, bb, 2)
	Equal(t, cc, 1)
}
