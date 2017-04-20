package pure

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
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
// NOTE: some stupid browsers send in locales lowercase when all the rest send it properly
func AcceptedLanguages(r *http.Request) (languages []string) {

	var accepted string

	if accepted = r.Header.Get(AcceptedLanguage); accepted == blank {
		return
	}

	options := strings.Split(accepted, ",")
	l := len(options)

	languages = make([]string, l)

	for i := 0; i < l; i++ {
		locale := strings.SplitN(options[i], ";", 2)
		languages[i] = strings.Trim(locale[0], " ")
	}

	return
}

// Attachment is a helper method for returning an attachement file
// to be downloaded, if you with to open inline see function Inline
func Attachment(w http.ResponseWriter, r io.Reader, filename string) (err error) {

	w.Header().Set(ContentDisposition, "attachment;filename="+filename)
	w.Header().Set(ContentType, detectContentType(filename))
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, r)

	return
}

// Inline is a helper method for returning a file inline to
// be rendered/opened by the browser
func Inline(w http.ResponseWriter, r io.Reader, filename string) (err error) {

	w.Header().Set(ContentDisposition, "inline;filename="+filename)
	w.Header().Set(ContentType, detectContentType(filename))
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, r)

	return
}

// ClientIP implements a best effort algorithm to return the real client IP, it parses
// X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
func ClientIP(r *http.Request) (clientIP string) {

	var values []string

	if values, _ = r.Header[XRealIP]; len(values) > 0 {

		clientIP = strings.TrimSpace(values[0])
		if clientIP != blank {
			return
		}
	}

	if values, _ = r.Header[XForwardedFor]; len(values) > 0 {
		clientIP = values[0]

		if index := strings.IndexByte(clientIP, ','); index >= 0 {
			clientIP = clientIP[0:index]
		}

		clientIP = strings.TrimSpace(clientIP)
		if clientIP != blank {
			return
		}
	}

	clientIP, _, _ = net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))

	return
}

// JSON marshals provided interface + returns JSON + status code
func JSON(w http.ResponseWriter, status int, i interface{}) error {

	b, err := json.Marshal(i)
	if err != nil {
		return err
	}

	w.Header().Set(ContentType, ApplicationJSONCharsetUTF8)
	w.WriteHeader(status)
	_, err = w.Write(b)

	return err
}

// JSONBytes returns provided JSON response with status code
func JSONBytes(w http.ResponseWriter, status int, b []byte) (err error) {

	w.Header().Set(ContentType, ApplicationJSONCharsetUTF8)
	w.WriteHeader(status)
	_, err = w.Write(b)
	return
}

// JSONP sends a JSONP response with status code and uses `callback` to construct
// the JSONP payload.
func JSONP(w http.ResponseWriter, status int, i interface{}, callback string) error {

	b, err := json.Marshal(i)
	if err != nil {
		return err
	}

	w.Header().Set(ContentType, ApplicationJavaScriptCharsetUTF8)
	w.WriteHeader(status)

	if _, err = w.Write([]byte(callback + "(")); err == nil {

		if _, err = w.Write(b); err == nil {
			_, err = w.Write([]byte(");"))
		}
	}

	return err
}

// XML marshals provided interface + returns XML + status code
func XML(w http.ResponseWriter, status int, i interface{}) error {

	b, err := xml.Marshal(i)
	if err != nil {
		return err
	}

	w.Header().Set(ContentType, ApplicationXMLCharsetUTF8)
	w.WriteHeader(status)

	if _, err = w.Write([]byte(xml.Header)); err == nil {
		_, err = w.Write(b)
	}

	return err
}

// XMLBytes returns provided XML response with status code
func XMLBytes(w http.ResponseWriter, status int, b []byte) (err error) {

	w.Header().Set(ContentType, ApplicationXMLCharsetUTF8)
	w.WriteHeader(status)

	if _, err = w.Write(xmlHeaderBytes); err == nil {
		_, err = w.Write(b)
	}

	return
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
// json.NewDecoder(io.LimitReader(r.Body, maxMemory)).Decode(v).
//
// NOTE: when includeQueryParams=true both query params and SEO query params will be parsed and
// included eg. route /user/:id?test=true both 'id' and 'test' are treated as query params and added
// to the request.Form prior to decoding or added to parsed JSON or XML; in short SEO query params are
// treated just like normal query params.
func Decode(r *http.Request, includeQueryParams bool, maxMemory int64, v interface{}) (err error) {

	typ := r.Header.Get(ContentType)

	if idx := strings.Index(typ, ";"); idx != -1 {
		typ = typ[:idx]
	}

	switch typ {

	case ApplicationJSON:
		err = DecodeJSON(r, includeQueryParams, maxMemory, v)

	case ApplicationXML:
		err = DecodeXML(r, includeQueryParams, maxMemory, v)

	case ApplicationForm:
		err = DecodeForm(r, includeQueryParams, v)

	case MultipartForm:
		err = DecodeMultipartForm(r, includeQueryParams, maxMemory, v)

	case ApplicationQueryParams:

		if includeQueryParams {
			err = DecodeQueryParams(r, includeQueryParams, v)
		}
	}

	return
}

// DecodeForm parses the requests form data into the provided struct.
//
// The Content-Type and http method are not checked.
//
// NOTE: when includeQueryParams=true both query params and SEO query params will be parsed and
// included eg. route /user/:id?test=true both 'id' and 'test' are treated as query params and added
// to the request.Form prior to decoding; in short SEO query params are treated just like normal query params.
func DecodeForm(r *http.Request, includeQueryParams bool, v interface{}) (err error) {

	if includeQueryParams {

		if err = ParseForm(r); err == nil {
			err = DefaultDecoder.Decode(v, r.Form)
		}

		return
	}

	if err = r.ParseForm(); err == nil {
		err = DefaultDecoder.Decode(v, r.PostForm)
	}

	return
}

// DecodeMultipartForm parses the requests form data into the provided struct.
//
// The Content-Type and http method are not checked.
//
// NOTE: when includeQueryParams=true both query params and SEO query params will be parsed and
// included eg. route /user/:id?test=true both 'id' and 'test' are treated as query params and added
// to the request.Form prior to decoding; in short SEO query params are treated just like normal query params.
func DecodeMultipartForm(r *http.Request, includeQueryParams bool, maxMemory int64, v interface{}) (err error) {

	if includeQueryParams {

		if err = ParseMultipartForm(r, maxMemory); err == nil {
			err = DefaultDecoder.Decode(v, r.Form)
		}

		return
	}

	if err = r.ParseMultipartForm(maxMemory); err == nil {
		err = DefaultDecoder.Decode(v, r.MultipartForm.Value)
	}

	return
}

// DecodeJSON decodes the request body into the provided struct and limits the request size via
// an io.LimitReader using the maxMemory param.
//
// The Content-Type e.g. "application/json" and http method are not checked.
//
// NOTE: when includeQueryParams=true both query params and SEO query params will be parsed and
// included eg. route /user/:id?test=true both 'id' and 'test' are treated as query params and
// added to parsed JSON; in short SEO query params are treated just like normal query params.
func DecodeJSON(r *http.Request, includeQueryParams bool, maxMemory int64, v interface{}) (err error) {

	err = json.NewDecoder(io.LimitReader(r.Body, maxMemory)).Decode(v)

	if includeQueryParams && err == nil {
		err = DecodeQueryParams(r, includeQueryParams, v)
	}

	return
}

// DecodeXML decodes the request body into the provided struct and limits the request size via
// an io.LimitReader using the maxMemory param.
//
// The Content-Type e.g. "application/xml" and http method are not checked.
//
// NOTE: when includeQueryParams=true both query params and SEO query params will be parsed and
// included eg. route /user/:id?test=true both 'id' and 'test' are treated as query params and
// added to parsed XML; in short SEO query params are treated just like normal query params.
func DecodeXML(r *http.Request, includeQueryParams bool, maxMemory int64, v interface{}) (err error) {

	err = xml.NewDecoder(io.LimitReader(r.Body, maxMemory)).Decode(v)

	if includeQueryParams && err == nil {
		err = DecodeQueryParams(r, includeQueryParams, v)
	}

	return
}

// DecodeQueryParams takes the URL Query params, adds SEO params or not based on the includeSEOQueryParams
// flag.
//
// NOTE: DecodeQueryParams is also used/called from Decode when no ContentType is specified
// the only difference is that it will always pass true for includeSEOQueryParams
func DecodeQueryParams(r *http.Request, includeSEOQueryParams bool, v interface{}) (err error) {
	err = DefaultDecoder.Decode(v, QueryParams(r, includeSEOQueryParams))
	return
}

// DecodeSEOQueryParams decodes the SEO Query params only and ignores the normal URL Query params.
func DecodeSEOQueryParams(r *http.Request, v interface{}) (err error) {

	if rvi := r.Context().Value(defaultContextIdentifier); rvi != nil {

		rv := rvi.(*requestVars)

		values := make(url.Values, len(rv.params))

		for _, p := range rv.params {
			values.Add(p.key, p.value)
		}

		err = DefaultDecoder.Decode(v, values)
	}

	return
}

// QueryParams returns the r.URL.Query() values and optionally have them include the
// SEO query params eg. route /users/:id?test=val if includeSEOQueryParams=true then
// values will include 'id' and 'test' values
func QueryParams(r *http.Request, includeSEOQueryParams bool) (values url.Values) {

	values = r.URL.Query()

	if includeSEOQueryParams {
		if rvi := r.Context().Value(defaultContextIdentifier); rvi != nil {

			rv := rvi.(*requestVars)

			for _, p := range rv.params {
				values.Add(p.key, p.value)
			}
		}
	}

	return
}
