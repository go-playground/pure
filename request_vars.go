package pure

import "context"

// ReqVars is the interface of request scoped variables
// tracked by pure
type ReqVars interface {
	URLParam(pname string) string
}

type requestVars struct {
	ctx        context.Context // holds a copy of it's parent requestVars
	params     urlParams
	formParsed bool
}

// Params returns the current routes Params
func (r *requestVars) URLParam(pname string) string {
	return r.params.Get(pname)
}
