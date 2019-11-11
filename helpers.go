package pure

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	httpext "github.com/go-playground/pkg/v4/net/http"
)

// RequestVars returns the request scoped variables tracked by pure
func RequestVars(r *http.Request) ReqVars {

	rv := r.Context().Value(defaultContextIdentifier)
	if rv == nil {
		return new(requestVars)
	}

	return rv.(*requestVars)
}

// AcceptedLanguages returns an array of accepted languages denoted by
// the Accept-Language header sent by the browser
func AcceptedLanguages(r *http.Request) (languages []string) {
	return httpext.AcceptedLanguages(r)
}

// Attachment is a helper method for returning an attachement file
// to be downloaded, if you with to open inline see function Inline
func Attachment(w http.ResponseWriter, r io.Reader, filename string) (err error) {
	return httpext.Attachment(w, r, filename)
}

// Inline is a helper method for returning a file inline to
// be rendered/opened by the browser
func Inline(w http.ResponseWriter, r io.Reader, filename string) (err error) {
	return httpext.Inline(w, r, filename)
}

// ClientIP implements a best effort algorithm to return the real client IP, it parses
// X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
func ClientIP(r *http.Request) (clientIP string) {
	return httpext.ClientIP(r)
}

// JSON marshals provided interface + returns JSON + status code
func JSON(w http.ResponseWriter, status int, i interface{}) error {
	return httpext.JSON(w, status, i)
}

// JSONBytes returns provided JSON response with status code
func JSONBytes(w http.ResponseWriter, status int, b []byte) (err error) {
	return httpext.JSONBytes(w, status, b)
}

// JSONP sends a JSONP response with status code and uses `callback` to construct
// the JSONP payload.
func JSONP(w http.ResponseWriter, status int, i interface{}, callback string) error {
	return httpext.JSONP(w, status, i, callback)
}

// XML marshals provided interface + returns XML + status code
func XML(w http.ResponseWriter, status int, i interface{}) error {
	return httpext.XML(w, status, i)
}

// XMLBytes returns provided XML response with status code
func XMLBytes(w http.ResponseWriter, status int, b []byte) (err error) {
	return httpext.XMLBytes(w, status, b)
}

// ParseForm calls the underlying http.Request ParseForm
// but also adds the URL params to the request Form as if
// they were defined as query params i.e. ?id=13&ok=true but
// does not add the params to the http.Request.URL.RawQuery
// for SEO purposes
func ParseForm(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	if rvi := r.Context().Value(defaultContextIdentifier); rvi != nil {
		rv := rvi.(*requestVars)
		if !rv.formParsed {
			for _, p := range rv.params {
				r.Form.Add(p.key, p.value)
			}
			rv.formParsed = true
		}
	}
	return nil
}

// ParseMultipartForm calls the underlying http.Request ParseMultipartForm
// but also adds the URL params to the request Form as if they were defined
// as query params i.e. ?id=13&ok=true but does not add the params to the
// http.Request.URL.RawQuery for SEO purposes
func ParseMultipartForm(r *http.Request, maxMemory int64) error {
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		return err
	}
	if rvi := r.Context().Value(defaultContextIdentifier); rvi != nil {
		rv := rvi.(*requestVars)
		if !rv.formParsed {
			for _, p := range rv.params {
				r.Form.Add(p.key, p.value)
			}
			rv.formParsed = true
		}
	}
	return nil
}

// Decode takes the request and attempts to discover it's content type via
// the http headers and then decode the request body into the provided struct.
// Example if header was "application/json" would decode using
// json.NewDecoder(ioext.LimitReader(r.Body, maxMemory)).Decode(v).
//
// NOTE: when qp=QueryParams both query params and SEO query params will be parsed and
// included eg. route /user/:id?test=true both 'id' and 'test' are treated as query params and added
// to the request.Form prior to decoding or added to parsed JSON or XML; in short SEO query params are
// treated just like normal query params.
func Decode(r *http.Request, qp httpext.QueryParamsOption, maxMemory int64, v interface{}) (err error) {
	typ := r.Header.Get(httpext.ContentType)
	if idx := strings.Index(typ, ";"); idx != -1 {
		typ = typ[:idx]
	}
	switch typ {
	case httpext.ApplicationForm:
		err = DecodeForm(r, qp, v)
	case httpext.MultipartForm:
		err = DecodeMultipartForm(r, qp, maxMemory, v)
	default:
		if qp == httpext.QueryParams {
			if err = DecodeSEOQueryParams(r, v); err != nil {
				return
			}
		}
		err = httpext.Decode(r, qp, maxMemory, v)
	}
	return
}

// DecodeForm parses the requests form data into the provided struct.
//
// The Content-Type and http method are not checked.
//
// NOTE: when qp=QueryParams both query params and SEO query params will be parsed and
// included eg. route /user/:id?test=true both 'id' and 'test' are treated as query params and added
// to the request.Form prior to decoding; in short SEO query params are treated just like normal query params.
func DecodeForm(r *http.Request, qp httpext.QueryParamsOption, v interface{}) (err error) {
	if qp == httpext.QueryParams {
		if err = ParseForm(r); err != nil {
			return
		}
	}
	err = httpext.DecodeForm(r, qp, v)
	return
}

// DecodeMultipartForm parses the requests form data into the provided struct.
//
// The Content-Type and http method are not checked.
//
// NOTE: when qp=QueryParams both query params and SEO query params will be parsed and
// included eg. route /user/:id?test=true both 'id' and 'test' are treated as query params and added
// to the request.Form prior to decoding; in short SEO query params are treated just like normal query params.
func DecodeMultipartForm(r *http.Request, qp httpext.QueryParamsOption, maxMemory int64, v interface{}) (err error) {
	if qp == httpext.QueryParams {
		if err = ParseMultipartForm(r, maxMemory); err != nil {
			return
		}
	}
	err = httpext.DecodeMultipartForm(r, qp, maxMemory, v)
	return
}

// DecodeJSON decodes the request body into the provided struct and limits the request size via
// an ioext.LimitReader using the maxMemory param.
//
// The Content-Type e.g. "application/json" and http method are not checked.
//
// NOTE: when qp=QueryParams both query params and SEO query params will be parsed and
// included eg. route /user/:id?test=true both 'id' and 'test' are treated as query params and
// added to parsed JSON; in short SEO query params are treated just like normal query params.
func DecodeJSON(r *http.Request, qp httpext.QueryParamsOption, maxMemory int64, v interface{}) error {
	return httpext.DecodeJSON(r, qp, maxMemory, v)
}

// DecodeXML decodes the request body into the provided struct and limits the request size via
// an ioext.LimitReader using the maxMemory param.
//
// The Content-Type e.g. "application/xml" and http method are not checked.
//
// NOTE: when qp=QueryParams both query params and SEO query params will be parsed and
// included eg. route /user/:id?test=true both 'id' and 'test' are treated as query params and
// added to parsed XML; in short SEO query params are treated just like normal query params.
func DecodeXML(r *http.Request, qp httpext.QueryParamsOption, maxMemory int64, v interface{}) error {
	return httpext.DecodeXML(r, qp, maxMemory, v)
}

// DecodeQueryParams takes the URL Query params, adds SEO params or not based on the includeSEOQueryParams
// flag.
//
// NOTE: DecodeQueryParams is also used/called from Decode when no ContentType is specified
// the only difference is that it will always decode SEO Query Params
func DecodeQueryParams(r *http.Request, qp httpext.QueryParamsOption, v interface{}) error {
	return httpext.DefaultFormDecoder.Decode(v, QueryParams(r, qp))
}

// DecodeSEOQueryParams decodes the SEO Query params only and ignores the normal URL Query params.
func DecodeSEOQueryParams(r *http.Request, v interface{}) (err error) {
	if rvi := r.Context().Value(defaultContextIdentifier); rvi != nil {
		rv := rvi.(*requestVars)
		values := make(url.Values, len(rv.params))
		for _, p := range rv.params {
			values.Add(p.key, p.value)
		}
		err = httpext.DefaultFormDecoder.Decode(v, values)
	}
	return
}

// QueryParams returns the r.URL.Query() values and optionally have them include the
// SEO query params eg. route /users/:id?test=val if qp=QueryParams then
// values will include 'id' as well as 'test' values
func QueryParams(r *http.Request, qp httpext.QueryParamsOption) (values url.Values) {
	values = r.URL.Query()
	if qp == httpext.QueryParams {
		if rvi := r.Context().Value(defaultContextIdentifier); rvi != nil {
			rv := rvi.(*requestVars)
			for _, p := range rv.params {
				values.Add(p.key, p.value)
			}
		}
	}
	return
}
