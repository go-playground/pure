package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-playground/pure"
)

func main() {

	p := pure.New()
	p.Use(logger)

	p.Get("/", helloWorld)

	http.ListenAndServe(":3007", p.Serve())
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

// logger middleware
func logger(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next(w, r)

		stop := time.Now()
		path := r.URL.Path

		if path == "" {
			path = "/"
		}

		log.Printf("%s %s %s", r.Method, path, stop.Sub(start))
	}
}
