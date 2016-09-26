package pure

import (
	"net/http"
	"net/url"
)

// ReqVars is the interface of request scoped variables
// tracked by pure
type ReqVars interface {
	URLParam(pname string) string
	QueryParams() url.Values
}

type requestVars struct {
	r           *http.Request
	params      Params
	queryParams url.Values
	formParsed  bool
}

// Params returns the current routes Params
func (r *requestVars) URLParam(pname string) string {
	return r.params.Get(pname)
}

// QueryParams returns the http.Request.URL Query Params
// reason for storing is if http.Request.URL.Query() is called
// more than once, it reparses all over again creating a new map and all.
// callin this will cache and on subsequent call will not reparse
func (r *requestVars) QueryParams() url.Values {

	if r.queryParams == nil {
		r.queryParams = r.r.URL.Query()
	}

	return r.queryParams
}
