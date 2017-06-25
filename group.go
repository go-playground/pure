package pure

import (
	"net/http"
	"strconv"
	"strings"
)

// IRouteGroup interface for router group
type IRouteGroup interface {
	IRoutes
	GroupWithNone(prefix string) IRouteGroup
	GroupWithMore(prefix string, middleware ...Middleware) IRouteGroup
	Group(prefix string) IRouteGroup
}

// IRoutes interface for routes
type IRoutes interface {
	Use(...Middleware)
	Any(string, http.HandlerFunc)
	Get(string, http.HandlerFunc)
	Post(string, http.HandlerFunc)
	Delete(string, http.HandlerFunc)
	Patch(string, http.HandlerFunc)
	Put(string, http.HandlerFunc)
	Options(string, http.HandlerFunc)
	Head(string, http.HandlerFunc)
	Connect(string, http.HandlerFunc)
	Trace(string, http.HandlerFunc)
}

// routeGroup struct containing all fields and methods for use.
type routeGroup struct {
	prefix     string
	middleware []Middleware
	pure       *Mux
}

var _ IRouteGroup = &routeGroup{}

func (g *routeGroup) handle(method string, path string, handler http.HandlerFunc) {

	if i := strings.Index(path, "//"); i != -1 {
		panic("Bad path '" + path + "' contains duplicate // at index:" + strconv.Itoa(i))
	}

	h := handler

	for i := len(g.middleware) - 1; i >= 0; i-- {
		h = g.middleware[i](h)
	}

	tree := g.pure.trees[method]

	if tree == nil {
		tree = new(node)
		g.pure.trees[method] = tree
	}

	pCount := tree.add(g.prefix+path, h)
	pCount++

	if pCount > g.pure.mostParams {
		g.pure.mostParams = pCount
	}
}

// Use adds a middleware handler to the group middleware chain.
func (g *routeGroup) Use(m ...Middleware) {
	g.middleware = append(g.middleware, m...)
}

// Connect adds a CONNECT route & handler to the router.
func (g *routeGroup) Connect(path string, h http.HandlerFunc) {
	g.handle(http.MethodConnect, path, h)
}

// Delete adds a DELETE route & handler to the router.
func (g *routeGroup) Delete(path string, h http.HandlerFunc) {
	g.handle(http.MethodDelete, path, h)
}

// Get adds a GET route & handler to the router.
func (g *routeGroup) Get(path string, h http.HandlerFunc) {
	g.handle(http.MethodGet, path, h)
}

// Head adds a HEAD route & handler to the router.
func (g *routeGroup) Head(path string, h http.HandlerFunc) {
	g.handle(http.MethodHead, path, h)
}

// Options adds an OPTIONS route & handler to the router.
func (g *routeGroup) Options(path string, h http.HandlerFunc) {
	g.handle(http.MethodOptions, path, h)
}

// Patch adds a PATCH route & handler to the router.
func (g *routeGroup) Patch(path string, h http.HandlerFunc) {
	g.handle(http.MethodPatch, path, h)
}

// Post adds a POST route & handler to the router.
func (g *routeGroup) Post(path string, h http.HandlerFunc) {
	g.handle(http.MethodPost, path, h)
}

// Put adds a PUT route & handler to the router.
func (g *routeGroup) Put(path string, h http.HandlerFunc) {
	g.handle(http.MethodPut, path, h)
}

// Trace adds a TRACE route & handler to the router.
func (g *routeGroup) Trace(path string, h http.HandlerFunc) {
	g.handle(http.MethodTrace, path, h)
}

// Handle allows for any method to be registered with the given
// route & handler. Allows for non standard methods to be used
// like CalDavs PROPFIND and so forth.
func (g *routeGroup) Handle(method string, path string, h http.HandlerFunc) {
	g.handle(method, path, h)
}

// Any adds a route & handler to the router for all HTTP methods.
func (g *routeGroup) Any(path string, h http.HandlerFunc) {
	g.Connect(path, h)
	g.Delete(path, h)
	g.Get(path, h)
	g.Head(path, h)
	g.Options(path, h)
	g.Patch(path, h)
	g.Post(path, h)
	g.Put(path, h)
	g.Trace(path, h)
}

// Match adds a route & handler to the router for multiple HTTP methods provided.
func (g *routeGroup) Match(methods []string, path string, h http.HandlerFunc) {
	for _, m := range methods {
		g.handle(m, path, h)
	}
}

// GroupWithNone creates a new sub router with specified prefix and no middleware attached.
func (g *routeGroup) GroupWithNone(prefix string) IRouteGroup {
	return &routeGroup{
		prefix:     g.prefix + prefix,
		pure:       g.pure,
		middleware: make([]Middleware, 0),
	}
}

// GroupWithMore creates a new sub router with specified prefix, retains existing middleware and adds new middleware.
func (g *routeGroup) GroupWithMore(prefix string, middleware ...Middleware) IRouteGroup {
	rg := &routeGroup{
		prefix:     g.prefix + prefix,
		pure:       g.pure,
		middleware: make([]Middleware, len(g.middleware)),
	}
	copy(rg.middleware, g.middleware)
	rg.Use(middleware...)
	return rg
}

// Group creates a new sub router with specified prefix and retains existing middleware.
func (g *routeGroup) Group(prefix string) IRouteGroup {
	rg := &routeGroup{
		prefix:     g.prefix + prefix,
		pure:       g.pure,
		middleware: make([]Middleware, len(g.middleware)),
	}
	copy(rg.middleware, g.middleware)
	return rg
}
