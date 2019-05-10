package main

import (
	"net/http"

	"github.com/go-playground/pure"
	mw "github.com/go-playground/pure/_examples/middleware/logging-recovery"
)

func main() {

	p := pure.New()
	p.Use(mw.LoggingAndRecovery(false))

	p.Get("/", helloWorld)

	http.ListenAndServe(":3007", p.Serve())
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}
