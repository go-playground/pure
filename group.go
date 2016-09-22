package pure

import (
	"net/http"
	"strconv"
	"strings"
)

// IRouteGroup interface for router group
type IRouteGroup interface {
	IRoutes
	Group(prefix string, middleware ...Middleware) IRouteGroup
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
	// WebSocket(websocket.Upgrader, string, HandlerFunc)
}

// routeGroup struct containing all fields and methods for use.
type routeGroup struct {
	prefix     string
	middleware []Middleware
	pure       *Pure
}

var _ IRouteGroup = &routeGroup{}

func (g *routeGroup) handle(method string, path string, handler http.HandlerFunc) {

	// if len(handlers) == 0 {
	// 	panic("No handler mapped to path:" + path)
	// }

	if i := strings.Index(path, "//"); i != -1 {
		panic("Bad path '" + path + "' contains duplicate // at index:" + strconv.Itoa(i))
	}

	h := handler

	for i := len(g.middleware) - 1; i >= 0; i-- {
		h = g.middleware[i](h)
	}

	var tree *node

	switch method {
	case http.MethodGet:
		tree = g.pure.get
	case http.MethodPost:
		tree = g.pure.post
	case http.MethodHead:
		tree = g.pure.head
	case http.MethodPut:
		tree = g.pure.put
	case http.MethodDelete:
		tree = g.pure.del
	case http.MethodConnect:
		tree = g.pure.connect
	case http.MethodOptions:
		tree = g.pure.options
	case http.MethodPatch:
		tree = g.pure.patch
	case http.MethodTrace:
		tree = g.pure.trace
	default:
		tree = g.pure.custom[method]
		if tree == nil {
			tree = new(node)
			g.pure.custom[method] = tree
		}
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

// // WebSocket adds a websocket route
// func (g *routeGroup) WebSocket(upgrader websocket.Upgrader, path string, h http.Handler) {

// 	// handler := g.lars.wrapHandler(h)
// 	g.Get(path, func(c Context) {

// 		ctx := c.BaseContext()
// 		var err error

// 		ctx.websocket, err = upgrader.Upgrade(ctx.response, ctx.request, nil)
// 		if err != nil {
// 			return
// 		}

// 		defer ctx.websocket.Close()
// 		c.Next()
// 	}, handler)
// }

// Group creates a new sub router with prefix. It inherits all properties from
// the parent. Passing middleware overrides parent middleware but still keeps
// the root level middleware intact.
func (g *routeGroup) Group(prefix string, middleware ...Middleware) IRouteGroup {

	rg := &routeGroup{
		prefix: g.prefix + prefix,
		pure:   g.pure,
	}

	if len(middleware) == 0 {
		rg.middleware = make([]Middleware, len(g.middleware)+len(middleware))
		copy(rg.middleware, g.middleware)

		return rg
	}

	if middleware[0] == nil {
		rg.middleware = make([]Middleware, 0)
		return rg
	}

	rg.middleware = make([]Middleware, len(middleware))
	copy(rg.middleware, g.pure.middleware)
	rg.Use(middleware...)

	return rg
}
