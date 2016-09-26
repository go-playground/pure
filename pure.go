package pure

import (
	"context"
	"encoding/xml"
	"net/http"
	"strings"
	"sync"
)

var (
	bpool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 64)
		},
	}

	defaultContextIdentifier = &struct {
		name string
	}{
		name: "pure",
	}

	xmlHeaderBytes = []byte(xml.Header)
)

// Mux is the main request multiplexer
type Mux struct {
	routeGroup
	get     *node
	post    *node
	del     *node
	put     *node
	head    *node
	connect *node
	options *node
	trace   *node
	patch   *node
	custom  map[string]*node

	// pool is used for reusable request scoped RequestVars content
	pool sync.Pool

	// mostParams used to keep track of the most amount of
	// params in any URL and this will set the default capacity
	// of each Params
	mostParams uint8

	http404     http.HandlerFunc // 404 Not Found
	http405     http.HandlerFunc // 405 Method Not Allowed
	httpOPTIONS http.HandlerFunc

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	redirectTrailingSlash bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	handleMethodNotAllowed bool

	// if enabled automatically handles OPTION requests; manually configured OPTION
	// handlers take presidence. default true
	automaticallyHandleOPTIONS bool
}

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// Get returns the URL parameter for the given key, or blank if not found
func (p Params) Get(key string) (param string) {

	for i := 0; i < len(p); i++ {
		if p[i].Key == key {
			param = p[i].Value
			return
		}
	}

	return
}

// Middleware is pure's middleware definition
type Middleware func(h http.HandlerFunc) http.HandlerFunc

var (
	default404Handler = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	methodNotAllowedHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	automaticOPTIONSHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
)

// New Creates and returns a new Pure instance
func New() *Mux {

	p := &Mux{
		routeGroup: routeGroup{
			middleware: make([]Middleware, 0),
		},
		get:                        new(node),
		post:                       new(node),
		del:                        new(node),
		put:                        new(node),
		head:                       new(node),
		connect:                    new(node),
		options:                    new(node),
		trace:                      new(node),
		patch:                      new(node),
		custom:                     make(map[string]*node),
		mostParams:                 0,
		http404:                    default404Handler,
		http405:                    methodNotAllowedHandler,
		httpOPTIONS:                automaticOPTIONSHandler,
		redirectTrailingSlash:      true,
		handleMethodNotAllowed:     false,
		automaticallyHandleOPTIONS: false,
	}

	p.routeGroup.pure = p
	p.pool.New = func() interface{} {

		return &requestVars{
			params: make(Params, p.mostParams),
		}
	}

	return p
}

// Register404 alows for overriding of the not found handler function.
// NOTE: this is run after not finding a route even after redirecting with the trailing slash
func (p *Mux) Register404(notFound http.HandlerFunc, middleware ...Middleware) {

	h := notFound

	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}

	p.http404 = h
}

// RegisterAutomaticOPTIONS tells pure whether to
// automatically handle OPTION requests; manually configured
// OPTION handlers take precedence. default true
func (p *Mux) RegisterAutomaticOPTIONS(middleware ...Middleware) {

	p.automaticallyHandleOPTIONS = true

	h := automaticOPTIONSHandler

	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}

	p.httpOPTIONS = h
}

// SetRedirectTrailingSlash tells pure whether to try
// and fix a URL by trying to find it
// lowercase -> with or without slash -> 404
func (p *Mux) SetRedirectTrailingSlash(set bool) {
	p.redirectTrailingSlash = set
}

// RegisterMethodNotAllowed tells pure whether to
// handle the http 405 Method Not Allowed status code
func (p *Mux) RegisterMethodNotAllowed(middleware ...Middleware) {

	p.handleMethodNotAllowed = true

	h := methodNotAllowedHandler

	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}

	p.http405 = h
}

// Serve returns an http.Handler to be used.
func (p *Mux) Serve() http.Handler {

	// reserved for any logic that needs to happen before serving starts.
	// i.e. although this router does not use priority to determine route order
	// could add sorting of tree nodes here....

	return http.HandlerFunc(p.serveHTTP)
}

// Conforms to the http.Handler interface.
func (p *Mux) serveHTTP(w http.ResponseWriter, r *http.Request) {

	var tree *node

	switch r.Method {
	case http.MethodGet:
		tree = p.get
	case http.MethodPost:
		tree = p.post
	case http.MethodHead:
		tree = p.head
	case http.MethodPut:
		tree = p.put
	case http.MethodDelete:
		tree = p.del
	case http.MethodConnect:
		tree = p.connect
	case http.MethodOptions:
		tree = p.options
	case http.MethodPatch:
		tree = p.patch
	case http.MethodTrace:
		tree = p.trace
	default:
		tree = p.custom[r.Method]
	}

	var h http.HandlerFunc
	rv := p.pool.Get().(*requestVars)
	rv.r = r

	if tree == nil {
		h = p.http404
		goto END
	}

	if h, rv.params = tree.find(r.URL.Path, rv.params[0:0]); h == nil {

		if p.redirectTrailingSlash && len(r.URL.Path) > 1 {

			// find again all lowercase
			orig := r.URL.Path
			lc := strings.ToLower(orig)

			if lc != r.URL.Path {

				if h, _ = tree.find(lc, rv.params[0:0]); h != nil {
					r.URL.Path = lc
					h = p.redirect(r.Method, r.URL.String())
					r.URL.Path = orig
					goto END
				}
			}

			if lc[len(lc)-1:] == basePath {
				lc = lc[:len(lc)-1]
			} else {
				lc = lc + basePath
			}

			if h, _ = tree.find(lc, rv.params[0:0]); h != nil {
				r.URL.Path = lc
				h = p.redirect(r.Method, r.URL.String())
				r.URL.Path = orig
				goto END
			}
		}

	} else {
		goto END
	}
	// }

	if p.automaticallyHandleOPTIONS && r.Method == http.MethodOptions {

		if r.URL.Path == "*" { // check server-wide OPTIONS

			if len(p.get.path) > 0 {
				w.Header().Add(Allow, http.MethodGet)
			}

			if len(p.post.path) > 0 {
				w.Header().Add(Allow, http.MethodPost)
			}

			if len(p.head.path) > 0 {
				w.Header().Add(Allow, http.MethodPut)
			}

			if len(p.put.path) > 0 {
				w.Header().Add(Allow, http.MethodPut)
			}

			if len(p.del.path) > 0 {
				w.Header().Add(Allow, http.MethodDelete)
			}

			if len(p.connect.path) > 0 {
				w.Header().Add(Allow, http.MethodConnect)
			}

			// if len(p.options.path) > 0 {
			// 	w.Header().Add(Allow, http.MethodOptions)
			// }

			if len(p.patch.path) > 0 {
				w.Header().Add(Allow, http.MethodPatch)
			}

			if len(p.trace.path) > 0 {
				w.Header().Add(Allow, http.MethodTrace)
			}

			for m := range p.custom {
				w.Header().Add(Allow, m)
			}

		} else {

			if len(p.get.path) > 0 && r.Method != http.MethodGet {
				if h, _ = p.get.find(r.URL.Path, rv.params[0:0]); h != nil {
					w.Header().Add(Allow, http.MethodGet)
				}
			}

			if len(p.post.path) > 0 && r.Method != http.MethodPost {
				if h, _ = p.post.find(r.URL.Path, rv.params[0:0]); h != nil {
					w.Header().Add(Allow, http.MethodPost)
				}
			}

			if len(p.head.path) > 0 && r.Method != http.MethodHead {
				if h, _ = p.head.find(r.URL.Path, rv.params[0:0]); h != nil {
					w.Header().Add(Allow, http.MethodHead)
				}
			}

			if len(p.put.path) > 0 && r.Method != http.MethodPut {
				if h, _ = p.put.find(r.URL.Path, rv.params[0:0]); h != nil {
					w.Header().Add(Allow, http.MethodPut)
				}
			}

			if len(p.del.path) > 0 && r.Method != http.MethodDelete {
				if h, _ = p.del.find(r.URL.Path, rv.params[0:0]); h != nil {
					w.Header().Add(Allow, http.MethodDelete)
				}
			}

			if len(p.connect.path) > 0 && r.Method != http.MethodConnect {
				if h, _ = p.connect.find(r.URL.Path, rv.params[0:0]); h != nil {
					w.Header().Add(Allow, http.MethodConnect)
				}
			}

			// options is a given, added below
			// if len(p.options.path) > 0 && r.Method != http.MethodOptions {
			// 	if h, _ = p.options.find(r.URL.Path, rv.params[0:0]); h != nil {
			// 		w.Header().Add(Allow, http.MethodOptions)
			// 	}
			// }

			if len(p.patch.path) > 0 && r.Method != http.MethodPatch {
				if h, _ = p.patch.find(r.URL.Path, rv.params[0:0]); h != nil {
					w.Header().Add(Allow, http.MethodPatch)
				}
			}

			if len(p.trace.path) > 0 && r.Method != http.MethodTrace {
				if h, _ = p.trace.find(r.URL.Path, rv.params[0:0]); h != nil {
					w.Header().Add(Allow, http.MethodTrace)
				}
			}

			for m, ctree := range p.custom {

				if h, _ = ctree.find(r.URL.Path, rv.params[0:0]); h != nil {
					w.Header().Add(Allow, m)
				}
			}
		}

		w.Header().Add(Allow, http.MethodOptions)
		h = p.httpOPTIONS

		goto END
	}

	if p.handleMethodNotAllowed {

		var found bool

		if len(p.get.path) > 0 && r.Method != http.MethodGet {
			if h, _ = p.get.find(r.URL.Path, rv.params[0:0]); h != nil {
				w.Header().Add(Allow, http.MethodGet)
				found = true
			}
		}

		if len(p.post.path) > 0 && r.Method != http.MethodPost {
			if h, _ = p.post.find(r.URL.Path, rv.params[0:0]); h != nil {
				w.Header().Add(Allow, http.MethodPost)
				found = true
			}
		}

		if len(p.head.path) > 0 && r.Method != http.MethodHead {
			if h, _ = p.head.find(r.URL.Path, rv.params[0:0]); h != nil {
				w.Header().Add(Allow, http.MethodHead)
				found = true
			}
		}

		if len(p.put.path) > 0 && r.Method != http.MethodPut {
			if h, _ = p.put.find(r.URL.Path, rv.params[0:0]); h != nil {
				w.Header().Add(Allow, http.MethodPut)
				found = true
			}
		}

		if len(p.del.path) > 0 && r.Method != http.MethodDelete {
			if h, _ = p.del.find(r.URL.Path, rv.params[0:0]); h != nil {
				w.Header().Add(Allow, http.MethodDelete)
				found = true
			}
		}

		if len(p.connect.path) > 0 && r.Method != http.MethodConnect {
			if h, _ = p.connect.find(r.URL.Path, rv.params[0:0]); h != nil {
				w.Header().Add(Allow, http.MethodConnect)
				found = true
			}
		}

		if len(p.options.path) > 0 && r.Method != http.MethodOptions {
			if h, _ = p.options.find(r.URL.Path, rv.params[0:0]); h != nil {
				w.Header().Add(Allow, http.MethodOptions)
				found = true
			}

		}

		if len(p.patch.path) > 0 && r.Method != http.MethodPatch {
			if h, _ = p.patch.find(r.URL.Path, rv.params[0:0]); h != nil {
				w.Header().Add(Allow, http.MethodPatch)
				found = true
			}
		}

		if len(p.trace.path) > 0 && r.Method != http.MethodTrace {
			if h, _ = p.trace.find(r.URL.Path, rv.params[0:0]); h != nil {
				w.Header().Add(Allow, http.MethodTrace)
				found = true
			}
		}

		for m, ctree := range p.custom {

			if m == r.Method {
				continue
			}

			if h, _ = ctree.find(r.URL.Path, rv.params[0:0]); h != nil {
				w.Header().Add(Allow, m)
				found = true
			}
		}

		if found {
			h = p.http405
			goto END
		}
	}

	// not found
	h = p.http404

END:

	if len(rv.params) > 0 {

		rv.formParsed = false

		// create requestVars and store on context
		r = r.WithContext(context.WithValue(r.Context(), defaultContextIdentifier, rv))
	}

	h(w, r)

	rv.queryParams = nil
	rv.r = nil
	p.pool.Put(rv)
}

func (p *Mux) redirect(method string, to string) (h http.HandlerFunc) {

	code := http.StatusMovedPermanently

	if method != http.MethodGet {
		code = http.StatusTemporaryRedirect
	}

	h = func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, code)
	}

	for i := len(p.middleware) - 1; i >= 0; i-- {
		h = p.middleware[i](h)
	}

	return
}
