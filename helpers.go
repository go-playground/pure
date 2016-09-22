package pure

import "net/http"

// // URLParams returns the params extracted from the URL
// // not to be confused with Query Parameters
// func URLParams(r *http.Request) Params {

// 	rv := r.Context().Value(defaultContextIdentifier).(*requestVars)
// 	if rv == nil {
// 		return nil
// 	}

// 	return rv.params
// }

// URLParam returns the params extracted from the URL
// not to be confused with Query Parameters
func URLParam(r *http.Request, pname string) string {

	return ""

	// rv := r.Context().Value(defaultContextIdentifier).(*requestVars)
	// if rv == nil {
	// 	return ""
	// }

	// // p := val.(*requestVars).params

	// for i := 0; i < len(rv.params); i++ {
	// 	if rv.params[i].Key == pname {
	// 		return rv.params[i].Key
	// 		// return
	// 	}
	// }

	// return ""
	// return val.(*requestVars).params.Get(pname)
}

// // QueryParams returns the http.Request.URL.Query() values
// // this function is not for convenience, but rather performance
// // URL.Query() reparses the RawQuery every time it's called, but this
// // function will cache the initial parsing so it doesn't have to reparse;
// // which is useful if when accessing these Params from multiple middleware.
// func QueryParams(r *http.Request, key int) Params {

// 	val := r.Context().Value(key)
// 	if val == nil {
// 		return nil
// 	}

// 	return val.(*requestVars).params
// }
