package pure

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
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

var (
	defaultHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Method))
	}

	idHandler = func(w http.ResponseWriter, r *http.Request) {
		rv := RequestVars(r)
		w.Write([]byte(rv.URLParam("id")))
	}

	params2Handler = func(w http.ResponseWriter, r *http.Request) {
		rv := RequestVars(r)
		w.Write([]byte(rv.URLParam("p1") + "|" + rv.URLParam("p2")))
	}

	defaultMiddleware = func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
		}
	}
)

func TestAllMethods(t *testing.T) {

	p := New()
	p.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
		}
	})

	tests := []struct {
		method  string
		path    string
		url     string
		handler http.HandlerFunc
		code    int
		body    string
		// panicExpected bool
		// panicMsg      string
	}{
		{
			method:  http.MethodGet,
			path:    "/get",
			url:     "/get",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodGet,
		},
		{
			method:  http.MethodPost,
			path:    "/post",
			url:     "/post",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodPost,
		},
		{
			method:  http.MethodHead,
			path:    "/head",
			url:     "/head",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodHead,
		},
		{
			method:  http.MethodPut,
			path:    "/put",
			url:     "/put",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodPut,
		},
		{
			method:  http.MethodDelete,
			path:    "/delete",
			url:     "/delete",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodDelete,
		},
		{
			method:  http.MethodConnect,
			path:    "/connect",
			url:     "/connect",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodConnect,
		},
		{
			method:  http.MethodOptions,
			path:    "/options",
			url:     "/options",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodOptions,
		},
		{
			method:  http.MethodPatch,
			path:    "/patch",
			url:     "/patch",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodPatch,
		},
		{
			method:  http.MethodTrace,
			path:    "/trace",
			url:     "/trace",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodTrace,
		},
		{
			method:  "PROPFIND",
			path:    "/propfind",
			url:     "/propfind",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    "PROPFIND",
		},
		{
			method:  http.MethodGet,
			path:    "/users/:id",
			url:     "/users/13",
			handler: idHandler,
			code:    http.StatusOK,
			body:    "13",
		},
		{
			method:  http.MethodGet,
			path:    "/2params/:p1",
			url:     "/2params/10",
			handler: params2Handler,
			code:    http.StatusOK,
			body:    "10|",
		},
		{
			method:  http.MethodGet,
			path:    "/2params/:p1/params/:p2",
			url:     "/2params/13/params/12",
			handler: params2Handler,
			code:    http.StatusOK,
			body:    "13|12",
		},
		{
			method:  http.MethodGet,
			path:    "/redirect",
			url:     "/redirect/",
			handler: defaultHandler,
			code:    http.StatusMovedPermanently,
			body:    "",
		},
		{
			method:  http.MethodPost,
			path:    "/redirect",
			url:     "/redirect/",
			handler: defaultHandler,
			code:    http.StatusTemporaryRedirect,
			body:    "",
		},
	}

	for _, tt := range tests {

		switch tt.method {
		case http.MethodGet:
			p.Get(tt.path, tt.handler)
		case http.MethodPost:
			p.Post(tt.path, tt.handler)
		case http.MethodHead:
			p.Head(tt.path, tt.handler)
		case http.MethodPut:
			p.Put(tt.path, tt.handler)
		case http.MethodDelete:
			p.Delete(tt.path, tt.handler)
		case http.MethodConnect:
			p.Connect(tt.path, tt.handler)
		case http.MethodOptions:
			p.Options(tt.path, tt.handler)
		case http.MethodPatch:
			p.Patch(tt.path, tt.handler)
		case http.MethodTrace:
			p.Trace(tt.path, tt.handler)
		default:
			p.Handle(tt.method, tt.path, tt.handler)
		}
	}

	hf := p.Serve()

	for _, tt := range tests {

		req, err := http.NewRequest(tt.method, tt.url, nil)
		if err != nil {
			t.Errorf("Expected 'nil' Got '%s'", err)
		}

		res := httptest.NewRecorder()

		hf.ServeHTTP(res, req)

		if res.Code != tt.code {
			t.Errorf("Expected '%d' Got '%d'", tt.code, res.Code)
		}

		if len(tt.body) > 0 {

			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Expected 'nil' Got '%s'", err)
			}

			s := string(b)

			if s != tt.body {
				t.Errorf("Expected '%s' Got '%s'", tt.body, s)
			}
		}
	}

	// test any

	p2 := New()
	p2.Any("/test", defaultHandler)

	hf = p2.Serve()

	test2 := []struct {
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

	for _, tt := range test2 {
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

func TestTooManyParams(t *testing.T) {

	s := "/"

	for i := 0; i < 256; i++ {
		s += ":id" + strconv.Itoa(i)
	}

	p := New()
	PanicMatches(t, func() { p.Get(s, defaultHandler) }, "too many parameters defined in path, max is 255")
}

func TestRouterAPI(t *testing.T) {
	p := New()

	for _, route := range githubAPI {
		p.handle(route.method, route.path, func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte(r.URL.Path)); err != nil {
				panic(err)
			}
		})
	}

	for _, route := range githubAPI {
		code, body := request(route.method, route.path, p)
		Equal(t, body, route.path)
		Equal(t, code, http.StatusOK)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	p := New()
	p.RegisterMethodNotAllowed(defaultMiddleware)

	p.Put("/home/", defaultHandler)
	p.Post("/home/", defaultHandler)
	p.Head("/home/", defaultHandler)
	p.Delete("/home/", defaultHandler)
	p.Connect("/home/", defaultHandler)
	p.Options("/home/", defaultHandler)
	p.Patch("/home/", defaultHandler)
	p.Trace("/home/", defaultHandler)
	p.Handle("PROPFIND", "/home/", defaultHandler)
	p.Handle("PROPFIND2", "/home/", defaultHandler)

	code, _ := request(http.MethodPut, "/home/", p)
	Equal(t, code, http.StatusOK)

	r, _ := http.NewRequest(http.MethodGet, "/home/", nil)
	w := httptest.NewRecorder()
	p.serveHTTP(w, r)

	Equal(t, w.Code, http.StatusMethodNotAllowed)

	allow, ok := w.Header()[Allow]
	Equal(t, ok, true)
	Equal(t, len(allow), 10)

	r, _ = http.NewRequest("PROPFIND2", "/home/1", nil)
	w = httptest.NewRecorder()
	p.serveHTTP(w, r)

	Equal(t, w.Code, http.StatusNotFound)
}

func TestMethodNotAllowed2(t *testing.T) {
	p := New()
	p.RegisterMethodNotAllowed()

	p.Get("/home/", defaultHandler)
	p.Head("/home/", defaultHandler)

	code, _ := request(http.MethodGet, "/home/", p)
	Equal(t, code, http.StatusOK)

	r, _ := http.NewRequest(http.MethodPost, "/home/", nil)
	w := httptest.NewRecorder()
	p.serveHTTP(w, r)

	Equal(t, w.Code, http.StatusMethodNotAllowed)

	allow, ok := w.Header()[Allow]

	// Sometimes this array is out of order for whatever reason?
	if allow[0] == http.MethodGet {
		Equal(t, ok, true)
		Equal(t, allow[0], http.MethodGet)
		Equal(t, allow[1], http.MethodHead)
	} else {
		Equal(t, ok, true)
		Equal(t, allow[1], http.MethodGet)
		Equal(t, allow[0], http.MethodHead)
	}
}

func TestAutomaticallyHandleOPTIONS(t *testing.T) {

	p := New()
	p.RegisterAutomaticOPTIONS(defaultMiddleware)
	p.Get("/home", defaultHandler)
	p.Post("/home", defaultHandler)
	p.Delete("/home", defaultHandler)
	p.Head("/home", defaultHandler)
	p.Put("/home", defaultHandler)
	p.Connect("/home", defaultHandler)
	p.Patch("/home", defaultHandler)
	p.Trace("/home", defaultHandler)
	p.Handle("PROPFIND", "/home", defaultHandler)

	code, _ := request(http.MethodGet, "/home", p)
	Equal(t, code, http.StatusOK)

	r, _ := http.NewRequest(http.MethodOptions, "/home", nil)
	w := httptest.NewRecorder()
	p.serveHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)

	allow, ok := w.Header()[Allow]

	Equal(t, ok, true)
	Equal(t, len(allow), 10)

	r, _ = http.NewRequest(http.MethodOptions, "*", nil)
	w = httptest.NewRecorder()
	p.serveHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)

	allow, ok = w.Header()[Allow]

	Equal(t, ok, true)
	Equal(t, len(allow), 10)
}

func TestRedirect(t *testing.T) {

	p := New()

	p.Get("/home/", defaultHandler)
	p.Post("/home/", defaultHandler)

	code, _ := request(http.MethodGet, "/home/", p)
	Equal(t, code, http.StatusOK)

	code, _ = request(http.MethodPost, "/home/", p)
	Equal(t, code, http.StatusOK)

	code, _ = request(http.MethodGet, "/home", p)
	Equal(t, code, http.StatusMovedPermanently)

	code, _ = request(http.MethodGet, "/Home/", p)
	Equal(t, code, http.StatusMovedPermanently)

	code, _ = request(http.MethodPost, "/home", p)
	Equal(t, code, http.StatusTemporaryRedirect)

	p.SetRedirectTrailingSlash(false)

	code, _ = request(http.MethodGet, "/home/", p)
	Equal(t, code, http.StatusOK)

	code, _ = request(http.MethodPost, "/home/", p)
	Equal(t, code, http.StatusOK)

	code, _ = request(http.MethodGet, "/home", p)
	Equal(t, code, http.StatusNotFound)

	code, _ = request(http.MethodGet, "/Home/", p)
	Equal(t, code, http.StatusNotFound)

	code, _ = request(http.MethodPost, "/home", p)
	Equal(t, code, http.StatusNotFound)

	p.SetRedirectTrailingSlash(true)

	p.Get("/users/:id", defaultHandler)
	p.Get("/users/:id/profile", defaultHandler)

	code, _ = request(http.MethodGet, "/users/10", p)
	Equal(t, code, http.StatusOK)

	code, _ = request(http.MethodGet, "/users/10/", p)
	Equal(t, code, http.StatusMovedPermanently)

	p.SetRedirectTrailingSlash(false)

	code, _ = request(http.MethodGet, "/users/10", p)
	Equal(t, code, http.StatusOK)

	code, _ = request(http.MethodGet, "/users/10/", p)
	Equal(t, code, http.StatusNotFound)
}

func TestNotFound(t *testing.T) {

	notFound := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	p := New()
	p.Register404(notFound, func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
		}
	})

	p.Get("/home/", defaultHandler)
	p.Post("/home/", defaultHandler)
	p.Get("/users/:id", defaultHandler)
	p.Get("/users/:id/:id2/:id3", defaultHandler)

	code, _ := request("BAD_METHOD", "/home/", p)
	Equal(t, code, http.StatusNotFound)

	code, _ = request(http.MethodGet, "/users/14/more", p)
	Equal(t, code, http.StatusNotFound)
}

func TestBadAdd(t *testing.T) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(r.Method)); err != nil {
			panic(err)
		}
	}

	p := New()
	PanicMatches(t, func() { p.Get("/%%%2frs#@$/", fn) }, "Query Unescape Error on path '/%%%2frs#@$/': invalid URL escape \"%%%\"")

	// bad existing params

	p.Get("/user/:id", fn)
	PanicMatches(t, func() { p.Get("/user/:user_id/profile", fn) }, "path segment ':user_id/profile' conflicts with existing wildcard ':id' in path '/user/:user_id/profile'")
	p.Get("/user/:id/profile", fn)

	p.Get("/admin/:id/profile", fn)
	PanicMatches(t, func() { p.Get("/admin/:admin_id", fn) }, "path segment ':admin_id' conflicts with existing wildcard ':id' in path '/admin/:admin_id'")

	PanicMatches(t, func() { p.Get("/assets/*/test", fn) }, "Character after the * symbol is not permitted, path '/assets/*/test'")

	p.Get("/superhero/*", fn)
	PanicMatches(t, func() { p.Get("/superhero/:id", fn) }, "path segment '/:id' conflicts with existing wildcard '/*' in path '/superhero/:id'")
	PanicMatches(t, func() { p.Get("/superhero/*", fn) }, "handlers are already registered for path '/superhero/*'")
	PanicMatches(t, func() { p.Get("/superhero/:id/", fn) }, "path segment '/:id/' conflicts with existing wildcard '/*' in path '/superhero/:id/'")

	p.Get("/supervillain/:id", fn)
	PanicMatches(t, func() { p.Get("/supervillain/*", fn) }, "path segment '*' conflicts with existing wildcard ':id' in path '/supervillain/*'")
	PanicMatches(t, func() { p.Get("/supervillain/:id", fn) }, "handlers are already registered for path '/supervillain/:id'")
}

func TestBasePath(t *testing.T) {

	p := New()
	p.Get("", defaultHandler)

	code, _ := request(http.MethodGet, "/", p)
	Equal(t, code, http.StatusOK)

}

type zombie struct {
	ID   int    `json:"id"   xml:"id"`
	Name string `json:"name" xml:"name"`
}

type route struct {
	method string
	path   string
}

var githubAPI = []route{
	// OAuth Authorizations
	{"GET", "/authorizations"},
	{"GET", "/authorizations/:id"},
	{"POST", "/authorizations"},
	//{"PUT", "/authorizations/clients/:client_id"},
	//{"PATCH", "/authorizations/:id"},
	{"DELETE", "/authorizations/:id"},
	{"GET", "/applications/:client_id/tokens/:access_token"},
	{"DELETE", "/applications/:client_id/tokens"},
	{"DELETE", "/applications/:client_id/tokens/:access_token"},

	// Activity
	{"GET", "/events"},
	{"GET", "/repos/:owner/:repo/events"},
	{"GET", "/networks/:owner/:repo/events"},
	{"GET", "/orgs/:org/events"},
	{"GET", "/users/:user/received_events"},
	{"GET", "/users/:user/received_events/public"},
	{"GET", "/users/:user/events"},
	{"GET", "/users/:user/events/public"},
	{"GET", "/users/:user/events/orgs/:org"},
	{"GET", "/feeds"},
	{"GET", "/notifications"},
	{"GET", "/repos/:owner/:repo/notifications"},
	{"PUT", "/notifications"},
	{"PUT", "/repos/:owner/:repo/notifications"},
	{"GET", "/notifications/threads/:id"},
	//{"PATCH", "/notifications/threads/:id"},
	{"GET", "/notifications/threads/:id/subscription"},
	{"PUT", "/notifications/threads/:id/subscription"},
	{"DELETE", "/notifications/threads/:id/subscription"},
	{"GET", "/repos/:owner/:repo/stargazers"},
	{"GET", "/users/:user/starred"},
	{"GET", "/user/starred"},
	{"GET", "/user/starred/:owner/:repo"},
	{"PUT", "/user/starred/:owner/:repo"},
	{"DELETE", "/user/starred/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/subscribers"},
	{"GET", "/users/:user/subscriptions"},
	{"GET", "/user/subscriptions"},
	{"GET", "/repos/:owner/:repo/subscription"},
	{"PUT", "/repos/:owner/:repo/subscription"},
	{"DELETE", "/repos/:owner/:repo/subscription"},
	{"GET", "/user/subscriptions/:owner/:repo"},
	{"PUT", "/user/subscriptions/:owner/:repo"},
	{"DELETE", "/user/subscriptions/:owner/:repo"},

	// Gists
	{"GET", "/users/:user/gists"},
	{"GET", "/gists"},
	//{"GET", "/gists/public"},
	//{"GET", "/gists/starred"},
	{"GET", "/gists/:id"},
	{"POST", "/gists"},
	//{"PATCH", "/gists/:id"},
	{"PUT", "/gists/:id/star"},
	{"DELETE", "/gists/:id/star"},
	{"GET", "/gists/:id/star"},
	{"POST", "/gists/:id/forks"},
	{"DELETE", "/gists/:id"},

	// Git Data
	{"GET", "/repos/:owner/:repo/git/blobs/:sha"},
	{"POST", "/repos/:owner/:repo/git/blobs"},
	{"GET", "/repos/:owner/:repo/git/commits/:sha"},
	{"POST", "/repos/:owner/:repo/git/commits"},
	//{"GET", "/repos/:owner/:repo/git/refs/*ref"},
	{"GET", "/repos/:owner/:repo/git/refs"},
	{"POST", "/repos/:owner/:repo/git/refs"},
	//{"PATCH", "/repos/:owner/:repo/git/refs/*ref"},
	//{"DELETE", "/repos/:owner/:repo/git/refs/*ref"},
	{"GET", "/repos/:owner/:repo/git/tags/:sha"},
	{"POST", "/repos/:owner/:repo/git/tags"},
	{"GET", "/repos/:owner/:repo/git/trees/:sha"},
	{"POST", "/repos/:owner/:repo/git/trees"},

	// Issues
	{"GET", "/issues"},
	{"GET", "/user/issues"},
	{"GET", "/orgs/:org/issues"},
	{"GET", "/repos/:owner/:repo/issues"},
	{"GET", "/repos/:owner/:repo/issues/:number"},
	{"POST", "/repos/:owner/:repo/issues"},
	//{"PATCH", "/repos/:owner/:repo/issues/:number"},
	{"GET", "/repos/:owner/:repo/assignees"},
	{"GET", "/repos/:owner/:repo/assignees/:assignee"},
	{"GET", "/repos/:owner/:repo/issues/:number/comments"},
	//{"GET", "/repos/:owner/:repo/issues/comments"},
	//{"GET", "/repos/:owner/:repo/issues/comments/:id"},
	{"POST", "/repos/:owner/:repo/issues/:number/comments"},
	//{"PATCH", "/repos/:owner/:repo/issues/comments/:id"},
	//{"DELETE", "/repos/:owner/:repo/issues/comments/:id"},
	{"GET", "/repos/:owner/:repo/issues/:number/events"},
	//{"GET", "/repos/:owner/:repo/issues/events"},
	//{"GET", "/repos/:owner/:repo/issues/events/:id"},
	{"GET", "/repos/:owner/:repo/labels"},
	{"GET", "/repos/:owner/:repo/labels/:name"},
	{"POST", "/repos/:owner/:repo/labels"},
	//{"PATCH", "/repos/:owner/:repo/labels/:name"},
	{"DELETE", "/repos/:owner/:repo/labels/:name"},
	{"GET", "/repos/:owner/:repo/issues/:number/labels"},
	{"POST", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name"},
	{"PUT", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones"},
	{"GET", "/repos/:owner/:repo/milestones/:number"},
	{"POST", "/repos/:owner/:repo/milestones"},
	//{"PATCH", "/repos/:owner/:repo/milestones/:number"},
	{"DELETE", "/repos/:owner/:repo/milestones/:number"},

	// Miscellaneous
	{"GET", "/emojis"},
	{"GET", "/gitignore/templates"},
	{"GET", "/gitignore/templates/:name"},
	{"POST", "/markdown"},
	{"POST", "/markdown/raw"},
	{"GET", "/meta"},
	{"GET", "/rate_limit"},

	// Organizations
	{"GET", "/users/:user/orgs"},
	{"GET", "/user/orgs"},
	{"GET", "/orgs/:org"},
	//{"PATCH", "/orgs/:org"},
	{"GET", "/orgs/:org/members"},
	{"GET", "/orgs/:org/members/:user"},
	{"DELETE", "/orgs/:org/members/:user"},
	{"GET", "/orgs/:org/public_members"},
	{"GET", "/orgs/:org/public_members/:user"},
	{"PUT", "/orgs/:org/public_members/:user"},
	{"DELETE", "/orgs/:org/public_members/:user"},
	{"GET", "/orgs/:org/teams"},
	{"GET", "/teams/:id"},
	{"POST", "/orgs/:org/teams"},
	//{"PATCH", "/teams/:id"},
	{"DELETE", "/teams/:id"},
	{"GET", "/teams/:id/members"},
	{"GET", "/teams/:id/members/:user"},
	{"PUT", "/teams/:id/members/:user"},
	{"DELETE", "/teams/:id/members/:user"},
	{"GET", "/teams/:id/repos"},
	{"GET", "/teams/:id/repos/:owner/:repo"},
	{"PUT", "/teams/:id/repos/:owner/:repo"},
	{"DELETE", "/teams/:id/repos/:owner/:repo"},
	{"GET", "/user/teams"},

	// Pull Requests
	{"GET", "/repos/:owner/:repo/pulls"},
	{"GET", "/repos/:owner/:repo/pulls/:number"},
	{"POST", "/repos/:owner/:repo/pulls"},
	//{"PATCH", "/repos/:owner/:repo/pulls/:number"},
	{"GET", "/repos/:owner/:repo/pulls/:number/commits"},
	{"GET", "/repos/:owner/:repo/pulls/:number/files"},
	{"GET", "/repos/:owner/:repo/pulls/:number/merge"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/merge"},
	{"GET", "/repos/:owner/:repo/pulls/:number/comments"},
	//{"GET", "/repos/:owner/:repo/pulls/comments"},
	//{"GET", "/repos/:owner/:repo/pulls/comments/:number"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/comments"},
	//{"PATCH", "/repos/:owner/:repo/pulls/comments/:number"},
	//{"DELETE", "/repos/:owner/:repo/pulls/comments/:number"},

	// Repositories
	{"GET", "/user/repos"},
	{"GET", "/users/:user/repos"},
	{"GET", "/orgs/:org/repos"},
	{"GET", "/repositories"},
	{"POST", "/user/repos"},
	{"POST", "/orgs/:org/repos"},
	{"GET", "/repos/:owner/:repo"},
	//{"PATCH", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/contributors"},
	{"GET", "/repos/:owner/:repo/languages"},
	{"GET", "/repos/:owner/:repo/teams"},
	{"GET", "/repos/:owner/:repo/tags"},
	{"GET", "/repos/:owner/:repo/branches"},
	{"GET", "/repos/:owner/:repo/branches/:branch"},
	{"DELETE", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/collaborators"},
	{"GET", "/repos/:owner/:repo/collaborators/:user"},
	{"PUT", "/repos/:owner/:repo/collaborators/:user"},
	{"DELETE", "/repos/:owner/:repo/collaborators/:user"},
	{"GET", "/repos/:owner/:repo/comments"},
	{"GET", "/repos/:owner/:repo/commits/:sha/comments"},
	{"POST", "/repos/:owner/:repo/commits/:sha/comments"},
	{"GET", "/repos/:owner/:repo/comments/:id"},
	//{"PATCH", "/repos/:owner/:repo/comments/:id"},
	{"DELETE", "/repos/:owner/:repo/comments/:id"},
	{"GET", "/repos/:owner/:repo/commits"},
	{"GET", "/repos/:owner/:repo/commits/:sha"},
	{"GET", "/repos/:owner/:repo/readme"},
	//{"GET", "/repos/:owner/:repo/contents/*path"},
	//{"PUT", "/repos/:owner/:repo/contents/*path"},
	//{"DELETE", "/repos/:owner/:repo/contents/*path"},
	//{"GET", "/repos/:owner/:repo/:archive_format/:ref"},
	{"GET", "/repos/:owner/:repo/keys"},
	{"GET", "/repos/:owner/:repo/keys/:id"},
	{"POST", "/repos/:owner/:repo/keys"},
	//{"PATCH", "/repos/:owner/:repo/keys/:id"},
	{"DELETE", "/repos/:owner/:repo/keys/:id"},
	{"GET", "/repos/:owner/:repo/downloads"},
	{"GET", "/repos/:owner/:repo/downloads/:id"},
	{"DELETE", "/repos/:owner/:repo/downloads/:id"},
	{"GET", "/repos/:owner/:repo/forks"},
	{"POST", "/repos/:owner/:repo/forks"},
	{"GET", "/repos/:owner/:repo/hooks"},
	{"GET", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks"},
	//{"PATCH", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks/:id/tests"},
	{"DELETE", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/merges"},
	{"GET", "/repos/:owner/:repo/releases"},
	{"GET", "/repos/:owner/:repo/releases/:id"},
	{"POST", "/repos/:owner/:repo/releases"},
	//{"PATCH", "/repos/:owner/:repo/releases/:id"},
	{"DELETE", "/repos/:owner/:repo/releases/:id"},
	{"GET", "/repos/:owner/:repo/releases/:id/assets"},
	{"GET", "/repos/:owner/:repo/stats/contributors"},
	{"GET", "/repos/:owner/:repo/stats/commit_activity"},
	{"GET", "/repos/:owner/:repo/stats/code_frequency"},
	{"GET", "/repos/:owner/:repo/stats/participation"},
	{"GET", "/repos/:owner/:repo/stats/punch_card"},
	{"GET", "/repos/:owner/:repo/statuses/:ref"},
	{"POST", "/repos/:owner/:repo/statuses/:ref"},

	// Search
	{"GET", "/search/repositories"},
	{"GET", "/search/code"},
	{"GET", "/search/issues"},
	{"GET", "/search/users"},
	{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword"},
	{"GET", "/legacy/repos/search/:keyword"},
	{"GET", "/legacy/user/search/:keyword"},
	{"GET", "/legacy/user/email/:email"},

	// Users
	{"GET", "/users/:user"},
	{"GET", "/user"},
	//{"PATCH", "/user"},
	{"GET", "/users"},
	{"GET", "/user/emails"},
	{"POST", "/user/emails"},
	{"DELETE", "/user/emails"},
	{"GET", "/users/:user/followers"},
	{"GET", "/user/followers"},
	{"GET", "/users/:user/following"},
	{"GET", "/user/following"},
	{"GET", "/user/following/:user"},
	{"GET", "/users/:user/following/:target_user"},
	{"PUT", "/user/following/:user"},
	{"DELETE", "/user/following/:user"},
	{"GET", "/users/:user/keys"},
	{"GET", "/user/keys"},
	{"GET", "/user/keys/:id"},
	{"POST", "/user/keys"},
	//{"PATCH", "/user/keys/:id"},
	{"DELETE", "/user/keys/:id"},
}

type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func (c *closeNotifyingRecorder) close() {
	c.closed <- true
}

func (c *closeNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}

func request(method, path string, p *Pure) (int, string) {
	r, _ := http.NewRequest(method, path, nil)
	w := &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
	hf := p.Serve()
	hf.ServeHTTP(w, r)
	return w.Code, w.Body.String()
}
