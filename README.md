package pure
============
<img align="right" src="https://raw.githubusercontent.com/go-playground/pure/master/logo.png">![Project status](https://img.shields.io/badge/version-5.0.0-green.svg)
[![Build Status](https://travis-ci.org/go-playground/pure.svg?branch=master)](https://travis-ci.org/go-playground/pure)
[![Coverage Status](https://coveralls.io/repos/github/go-playground/pure/badge.svg?branch=master)](https://coveralls.io/github/go-playground/pure?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-playground/pure)](https://goreportcard.com/report/github.com/go-playground/pure)
[![GoDoc](https://godoc.org/github.com/go-playground/pure?status.svg)](https://godoc.org/github.com/go-playground/pure)
![License](https://img.shields.io/dub/l/vibe-d.svg)
[![Gitter](https://badges.gitter.im/go-playground/pure.svg)](https://gitter.im/go-playground/pure?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

Pure is a fast radix-tree based HTTP router that sticks to the native implementations of Go's "net/http" package;
in essence, keeping the handler implementations 'pure' by using Go 1.7's "context" package.

This makes heavy usage of `github.com/go-playground/pkg` for HTTP abstractions.

Why Another HTTP Router?
------------------------
I initially created [lars](https://github.com/go-playground/lars), which I still maintain, that wraps the native implementation, think of this package as a Go pure implementation of [lars](https://github.com/go-playground/lars)

Key & Unique Features 
--------------
- [x] It sticks to Go's native implementations while providing helper functions for convenience
- [x] **Fast & Efficient** - pure uses a custom version of [httprouter](https://github.com/julienschmidt/httprouter)'s radix tree, so incredibly fast and efficient.

Installation
-----------

Use go get 

```shell
go get -u github.com/go-playground/pure
```

Usage
------
```go
package main

import (
	"net/http"

	"github.com/go-playground/pure"
	mw "github.com/go-playground/pure/_examples/middleware/logging-recovery"
)

func main() {

	p := pure.New()
	p.Use(mw.LoggingAndRecovery(true))

	p.Get("/", helloWorld)

	http.ListenAndServe(":3007", p.Serve())
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}
```

RequestVars
-----------
This is an interface that is used to pass request scoped variables and functions using `context.Context`.
It is implemented in this way because retrieving values from `context` isn't the fastest, and so using this 
the router can store multiple pieces of information while reducing lookup time to a single stored `RequestVars`.

Currently only the URL/SEO params are stored on the `RequestVars` but if/when more is added they can merely be added
to the `RequestVars` and there will be no additional lookup time.

URL Params
----------

```go
p := p.New()

// the matching param will be stored in the context's params with name "id"
p.Get("/user/:id", UserHandler)

// extract params like so
rv := pure.RequestVars(r) // done this way so only have to extract from context once, read above
rv.URLParam(paramname)

// serve css, js etc.. pure.RequestVars(r).URLParam(pure.WildcardParam) will return the remaining path if 
// you need to use it in a custom handler...
p.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP)

...
```

**Note:** Since this router has only explicit matches, you can not register static routes and parameters for the same path segment. For example you can not register the patterns /user/new and /user/:user for the same request method at the same time. The routing of different request methods is independent from each other. I was initially against this, however it nearly cost me in a large web application where the dynamic param value say :type actually could have matched another static route and that's just too dangerous and so it is not allowed.

Groups
-----
```go

p.Use(LoggingAndRecovery, Gzip...)
...
p.Post("/users/add", ...)

// creates a group for /user/:userid + inherits all middleware registered previously by p
user := p.Group("/user/:userid")
user.Get("", ...)
user.Post("", ...)
user.Delete("/delete", ...)

contactInfo := user.Group("/contact-info/:cid")
contactinfo.Delete("/delete", ...)

// creates a group for /others, inherits all middleware registered previously by p + adds 
// OtherHandler to middleware
others := p.GroupWithMore("/others", OtherHandler)

// creates a group for /admin WITH NO MIDDLEWARE... more can be added using admin.Use()
admin := p.GroupWithNone("/admin")
admin.Use(SomeAdminSecurityMiddleware)
...
```

Decoding Body
-------------
currently JSON, XML, FORM, Multipart Form and url.Values are support out of the box; there are also 
individual functions for each as well when you know the Content-Type.
```go
	// second argument denotes yes or no I would like URL query parameter fields
	// to be included. i.e. 'id' and 'id2' in route '/user/:id?id2=val' should it be included.
	if err := pure.Decode(r, true, maxBytes, &user); err != nil {
		log.Println(err)
	}
```

Misc
-----
```go

// set custom 404 ( not Found ) handler
p.Register404(404Handler, middleware_like_logging)

// Redirect to or from ending slash if route not found, default is true
p.SetRedirectTrailingSlash(true)

// Handle 405 ( Method Not allowed ), default is false
p.RegisterMethodNotAllowed(middleware)

// automatically handle OPTION requests; manually configured
// OPTION handlers take precedence. default false
p.RegisterAutomaticOPTIONS(middleware)

```

Middleware
-----------
There are some pre-defined middlewares within the middleware folder; NOTE: that the middleware inside will
comply with the following rule(s):

* Are completely reusable by the community without modification

Other middleware will be listed under the _examples/middleware/... folder for a quick copy/paste modify. As an example a LoddingAndRecovery middleware is very application dependent and therefore will be listed under the _examples/middleware/...

Benchmarks
-----------
Run on i5-7600 16 GB DDR4-2400 using Go version go1.12.5 darwin/amd64

NOTICE: pure uses a custom version of [httprouter](https://github.com/julienschmidt/httprouter)'s radix tree, benchmarks can be found [here](https://github.com/deankarn/go-http-routing-benchmark/tree/pure-and-lars) the slowdown is with the use of the `context` package, as you can see when no SEO params are defined, and therefore no need to store anything in the context, it is faster than even lars.

```go
go test -bench=. -benchmem=true
#GithubAPI Routes: 203
   Pure: 37096 Bytes

#GPlusAPI Routes: 13
   Pure: 2792 Bytes

#ParseAPI Routes: 26
   Pure: 5040 Bytes

#Static Routes: 157
   HttpServeMux: 14992 Bytes
   Pure: 21096 Bytes


BenchmarkPure_Param             10000000               184 ns/op             384 B/op          2 allocs/op
BenchmarkPure_Param5            10000000               236 ns/op             384 B/op          2 allocs/op
BenchmarkPure_Param20            5000000               393 ns/op             384 B/op          2 allocs/op
BenchmarkPure_ParamWrite         5000000               240 ns/op             384 B/op          2 allocs/op
BenchmarkPure_GithubStatic      50000000                36.2 ns/op             0 B/op          0 allocs/op
BenchmarkPureGithubParam        10000000               230 ns/op             384 B/op          2 allocs/op
BenchmarkPure_GithubAll            30000             43887 ns/op           64130 B/op        334 allocs/op
BenchmarkPure_GPlusStatic       50000000                22.8 ns/op             0 B/op          0 allocs/op
BenchmarkPure_GPlusParam        10000000               192 ns/op             384 B/op          2 allocs/op
BenchmarkPure_GPlus2Params      10000000               211 ns/op             384 B/op          2 allocs/op
BenchmarkPure_GPlusAll            500000              2457 ns/op            4224 B/op         22 allocs/op
BenchmarkPure_ParseStatic       100000000               23.7 ns/op             0 B/op          0 allocs/op
BenchmarkPure_ParseParam        10000000               177 ns/op             384 B/op          2 allocs/op
BenchmarkPure_Parse2Params      10000000               193 ns/op             384 B/op          2 allocs/op
BenchmarkPure_ParseAll            500000              3751 ns/op            6144 B/op         32 allocs/op
BenchmarkPure_StaticAll           200000              8574 ns/op               0 B/op          0 allocs/op
```

Package Versioning
----------
I'm jumping on the vendoring bandwagon, you should vendor this package as I will not
be creating different version with gopkg.in like allot of my other libraries.

Why? because my time is spread pretty thin maintaining all of the libraries I have + LIFE,
it is so freeing not to worry about it and will help me keep pouring out bigger and better
things for you the community.

I am open versioning with gopkg.in should anyone request it, but this should be stable going forward.

Licenses
--------
- [MIT License](https://raw.githubusercontent.com/go-playground/pure/master/LICENSE) (MIT), Copyright (c) 2016 Dean Karn
- [BSD License](https://raw.githubusercontent.com/julienschmidt/httprouter/master/LICENSE), Copyright (c) 2013 Julien Schmidt. All rights reserved.
