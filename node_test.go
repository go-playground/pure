package pure

import (
	"net/http"
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

func TestAddChain(t *testing.T) {
	p := New()

	p.Get("/home", defaultHandler)

	PanicMatches(t, func() { p.Get("/home", defaultHandler) }, "handlers are already registered for path '/home'")
}

func TestBadWildcard(t *testing.T) {

	p := New()
	PanicMatches(t, func() { p.Get("/test/:test*test", defaultHandler) }, "only one wildcard per path segment is allowed, has: ':test*test' in path '/test/:test*test'")

	p.Get("/users/:id/contact-info/:cid", defaultHandler)
	PanicMatches(t, func() { p.Get("/users/:id/*", defaultHandler) }, "wildcard route '*' conflicts with existing children in path '/users/:id/*'")
	PanicMatches(t, func() { p.Get("/admin/:/", defaultHandler) }, "wildcards must be named with a non-empty name in path '/admin/:/'")
	PanicMatches(t, func() { p.Get("/admin/events*", defaultHandler) }, "no / before catch-all in path '/admin/events*'")

	l2 := New()
	l2.Get("/", defaultHandler)
	PanicMatches(t, func() { l2.Get("/*", defaultHandler) }, "catch-all conflicts with existing handle for the path segment root in path '/*'")

	code, _ := request(http.MethodGet, "/home", l2)
	Equal(t, code, http.StatusNotFound)

	l3 := New()
	l3.Get("/testers/:id", defaultHandler)

	code, _ = request(http.MethodGet, "/testers/13/test", l3)
	Equal(t, code, http.StatusNotFound)
}

func TestDuplicateParams(t *testing.T) {

	p := New()
	p.Get("/store/:id", defaultHandler)
	PanicMatches(t, func() { p.Get("/store/:id/employee/:id", defaultHandler) }, "Duplicate param name ':id' detected for route '/store/:id/employee/:id'")

	p.Get("/company/:id/", defaultHandler)
	PanicMatches(t, func() { p.Get("/company/:id/employee/:id/", defaultHandler) }, "Duplicate param name ':id' detected for route '/company/:id/employee/:id/'")
}

func TestWildcardParam(t *testing.T) {
	p := New()
	p.Get("/users/*", func(w http.ResponseWriter, r *http.Request) {

		rv := RequestVars(r)
		if _, err := w.Write([]byte(rv.URLParam(WildcardParam))); err != nil {
			panic(err)
		}
	})

	code, body := request(http.MethodGet, "/users/testwild", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "testwild")

	code, body = request(http.MethodGet, "/users/testwildslash/", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "testwildslash/")
}

func TestBadRoutes(t *testing.T) {
	p := New()

	PanicMatches(t, func() { p.Get("/users//:id", defaultHandler) }, "Bad path '/users//:id' contains duplicate // at index:6")
}
