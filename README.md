##Pure
<img align="right" src="https://raw.githubusercontent.com/go-playground/pure/master/logo.png">
![Project status](https://img.shields.io/badge/version-2.4.0-green.svg)
[![Build Status](https://semaphoreci.com/api/v1/joeybloggs/pure/branches/master/badge.svg)](https://semaphoreci.com/joeybloggs/pure)
[![Coverage Status](https://coveralls.io/repos/github/go-playground/pure/badge.svg?branch=master)](https://coveralls.io/github/go-playground/pure?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-playground/pure)](https://goreportcard.com/report/github.com/go-playground/pure)
[![GoDoc](https://godoc.org/github.com/go-playground/pure?status.svg)](https://godoc.org/github.com/go-playground/pure)
![License](https://img.shields.io/dub/l/vibe-d.svg)
[![Gitter](https://badges.gitter.im/go-playground/pure.svg)](https://gitter.im/go-playground/pure?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

Pure is a fast radix-tree based HTTP router that sticks to the native implimentations of Go's "net/http" package;
in essence, keeping the handler implimentations 'pure' by using Go 1.7's "context" package.

Why Another HTTP Router?
------------------------
I initially created [lars](https://github.com/go-playground/lars), which I still maintain, that wraps the native implimentation, think of this package as a Go pure implimentation of [lars](https://github.com/go-playground/lars)

Key & Unique Features 
--------------
- [x] It sticks to Go's native implimentations while providing helper functions for convenience
- [x] **Fast & Efficient** - pure uses a custom version of [httprouter](https://github.com/julienschmidt/httprouter) so incredibly fast and efficient.

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
	mw "github.com/go-playground/pure/examples/middleware/logging-recovery"
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

URL Params
----------

```go
p := p.New()

// the matching param will be stored in the context's params with name "id"
l.Get("/user/:id", UserHandler)

// extract params like so
rv := pure.ReqestVars(r) // done this way so only have to extract from context once
rv.Params().Get(paramname)

// serve css, js etc.. pure.RequestVars(r).Params().Get(pure.WildcardParam) will return the remaining path if 
// you need to use it in a custom handler...
l.Get("/static/*", http.FileServer(http.Dir("static/"))) 

...
```

**Note:** Since this router has only explicit matches, you can not register static routes and parameters for the same path segment. For example you can not register the patterns /user/new and /user/:user for the same request method at the same time. The routing of different request methods is independent from each other. I was initially against this, and this router allowed it in a previous version, however it nearly cost me in a big app where the dynamic param value say :type actually could have matched another static route and that's just too dangerous, so it is no longer allowed.

Groups
-----
```go

p.Use(LoggingAndRecovery)
...
p.Post("/users/add", ...)

// creates a group for user + inherits all middleware registered using p.Use()
user := p.Group("/user/:userid")
user.Get("", ...)
user.Post("", ...)
user.Delete("/delete", ...)

contactInfo := user.Group("/contact-info/:ciid")
contactinfo.Delete("/delete", ...)

// creates a group for others + inherits all middleware registered using p.Use() + adds 
// OtherHandler to middleware
others := p.Group("/others", OtherHandler)

// creates a group for admin WITH NO MIDDLEWARE... more can be added using admin.Use()
admin := p.Group("/admin",nil)
admin.Use(SomeAdminSecurityMiddleware)
...
```

Decoding Body
-------------
currently JSON, XML, FORM + Multipart Form's are support out of the box.
```go
	// second argument denotes yes or no I would like URL query parameter fields
	// to be included. i.e. 'id' in route '/user?id=val' should it be included.
	if err := pure.Decode(r, true, maxBytes, &user); err != nil {
		log.Println(err)
	}
```

Misc
-----
```go

// set custom 404 ( not Found ) handler
l.Register404(404Handler, middleware_like_logging)

// Redirect to or from ending slash if route not found, default is true
l.SetRedirectTrailingSlash(true)

// Handle 405 ( Method Not allowed ), default is false
l.RegisterMethodNotAllowed(middleware)

// automatically handle OPTION requests; manually configured
// OPTION handlers take precedence. default false
l.RegisterAutomaticOPTIONS(middleware)

```

Middleware
-----------
There are some pre-defined middlewares within the middleware folder; NOTE: that the middleware inside will
comply with the following rule(s):

* Are completely reusable by the community without modification

Other middleware will be listed under the examples/middleware/... folder for a quick copy/paste modify. as an example a logging or
recovery middleware are very application dependent and therefore will be listed under the examples/middleware/...

Benchmarks
-----------
Run on MacBook Pro (Retina, 15-inch, Late 2013) 2.6 GHz Intel Core i7 16 GB 1600 MHz DDR3 using Go version go1.7.1 darwin/amd64

NOTICE: pure uses a custom version of [httprouter](https://github.com/julienschmidt/httprouter), benchmarks can be found [here](https://github.com/joeybloggs/go-http-routing-benchmark/tree/pure-and-lars)
the slowdown is with the use of the `context` package, as you can see when no params, and therefore no need to store anything in the context, it is faster than even lars.

```go
go test -bench=. -benchmem=true
#GithubAPI Routes: 203
   Pure: 37560 Bytes

#GPlusAPI Routes: 13
   Pure: 2808 Bytes

#ParseAPI Routes: 26
   Pure: 5072 Bytes

#Static Routes: 157
   Pure: 21224 Bytes

BenchmarkPure_Param        	10000000	       157 ns/op	     240 B/op	       1 allocs/op
BenchmarkPure_Param5       	10000000	       208 ns/op	     240 B/op	       1 allocs/op
BenchmarkPure_Param20      	 5000000	       350 ns/op	     240 B/op	       1 allocs/op
BenchmarkPure_ParamWrite   	10000000	       221 ns/op	     240 B/op	       1 allocs/op
BenchmarkPure_GithubStatic 	20000000	        72.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkPure_GithubParam  	10000000	       230 ns/op	     240 B/op	       1 allocs/op
BenchmarkPure_GithubAll    	   30000	     43054 ns/op	   40082 B/op	     167 allocs/op
BenchmarkPure_GPlusStatic  	30000000	        54.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkPure_GPlusParam   	10000000	       182 ns/op	     240 B/op	       1 allocs/op
BenchmarkPure_GPlus2Params 	10000000	       207 ns/op	     240 B/op	       1 allocs/op
BenchmarkPure_GPlusAll     	 1000000	      2297 ns/op	    2640 B/op	      11 allocs/op
BenchmarkPure_ParseStatic  	30000000	        56.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkPure_ParseParam   	10000000	       166 ns/op	     240 B/op	       1 allocs/op
BenchmarkPure_Parse2Params 	10000000	       180 ns/op	     240 B/op	       1 allocs/op
BenchmarkPure_ParseAll     	  500000	      3671 ns/op	    3840 B/op	      16 allocs/op
BenchmarkPure_StaticAll    	  100000	     14646 ns/op	       0 B/op	       0 allocs/op
```

Package Versioning
----------
I'm jumping on the vendoring bandwagon, you should vendor this package as I will not
be creating different version with gopkg.in like allot of my other libraries.

Why? because my time is spread pretty thin maintaining all of the libraries I have + LIFE,
it is so freeing not to worry about it and will help me keep pouring out bigger and better
things for you the community.

Licenses
--------
- [MIT License](https://raw.githubusercontent.com/go-playground/pure/master/LICENSE) (MIT), Copyright (c) 2016 Dean Karn
- [BSD License](https://raw.githubusercontent.com/julienschmidt/httprouter/master/LICENSE), Copyright (c) 2013 Julien Schmidt. All rights reserved.
