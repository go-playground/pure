package pure

// // RequestVars is the type stored in the http.Request's Context
// // for Request scoped variables tracked/stored by the pure router
// type RequestVars interface {
// 	Params() Params
// 	QueryParams() url.Values
// }

var defaultContextIdentifier = &struct {
	name string
}{
	name: "pure",
}

type requestVars struct {
	params Params
	// queryParams url.Values
}

// // Params returns the current routes Params
// func (r *requestVars) Params() Params {
// 	return r.params
// }

// // QueryParams returns the http.Request.URL Query Params
// // reason for storing is if http.Request.URL.Query() is called
// // more than once, it reparses all over again creating a new map and all.
// // callin this will cache and on subsequent call will not reparse
// func (r *requestVars) QueryParams() url.Values {
// 	return r.queryParams
// }
