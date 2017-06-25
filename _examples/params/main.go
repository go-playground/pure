package main

import (
	"net/http"

	"github.com/go-playground/pure"
	mw "github.com/go-playground/pure/_examples/middleware/logging-recovery"
)

func main() {

	p := pure.New()
	p.Use(mw.LoggingAndRecovery(true))

	p.Get("/user/:id", user)

	http.ListenAndServe(":3007", p.Serve())
}

func user(w http.ResponseWriter, r *http.Request) {
	rv := pure.RequestVars(r)

	w.Write([]byte("USER_ID:" + rv.URLParam("id")))
}
