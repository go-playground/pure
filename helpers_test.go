package pure

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	. "gopkg.in/go-playground/assert.v1"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//

func TestNoRequestVars(t *testing.T) {

	reqVars := func(w http.ResponseWriter, r *http.Request) {
		RequestVars(r)
	}

	p := New()
	p.Get("/home", reqVars)

	code, _ := request(http.MethodGet, "/home", p)
	Equal(t, code, http.StatusOK)
}

func TestQueryParams(t *testing.T) {

	qp1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rv := RequestVars(r)
			w.Write([]byte(rv.QueryParams().Get("test")))
			next(w, r)
		}
	}

	qp2 := func(w http.ResponseWriter, r *http.Request) {
		rv := RequestVars(r)
		w.Write([]byte(rv.QueryParams().Get("test")))
	}

	p := New()
	p.Use(qp1)
	p.Get("/home/:test", qp2)

	code, body := request(http.MethodGet, "/home/testval?test=val", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "valval")
}

func TestDecode(t *testing.T) {

	type TestStruct struct {
		ID              int `form:"id"`
		Posted          string
		MultiPartPosted string
	}

	test := new(TestStruct)

	p := New()
	p.Post("/decode/:id", func(w http.ResponseWriter, r *http.Request) {
		err := Decode(r, true, 16<<10, test)
		Equal(t, err, nil)
	})
	p.Post("/decode2/:id", func(w http.ResponseWriter, r *http.Request) {
		err := Decode(r, false, 16<<10, test)
		Equal(t, err, nil)
	})
	p.Post("/decode3/:id", func(w http.ResponseWriter, r *http.Request) {
		err := Decode(r, true, 16<<10, test)
		Equal(t, err, nil)
	})

	hf := p.Serve()

	form := url.Values{}
	form.Add("Posted", "value")

	r, _ := http.NewRequest(http.MethodPost, "/decode/14?id=13", strings.NewReader(form.Encode()))
	r.Header.Set(ContentType, ApplicationForm)
	w := httptest.NewRecorder()

	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 13)
	Equal(t, test.Posted, "value")
	Equal(t, test.MultiPartPosted, "")

	r, _ = http.NewRequest(http.MethodPost, "/decode/14", strings.NewReader(form.Encode()))
	r.Header.Set(ContentType, ApplicationForm)
	w = httptest.NewRecorder()

	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 14)
	Equal(t, test.Posted, "value")
	Equal(t, test.MultiPartPosted, "")

	test = new(TestStruct)
	r, _ = http.NewRequest(http.MethodPost, "/decode2/13", strings.NewReader(form.Encode()))
	r.Header.Set(ContentType, ApplicationForm)
	w = httptest.NewRecorder()

	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 0)
	Equal(t, test.Posted, "value")
	Equal(t, test.MultiPartPosted, "")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err := writer.WriteField("MultiPartPosted", "value")
	Equal(t, err, nil)

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	err = writer.Close()
	Equal(t, err, nil)

	test = new(TestStruct)
	r, _ = http.NewRequest(http.MethodPost, "/decode/13?id=12", body)
	r.Header.Set(ContentType, writer.FormDataContentType())
	w = httptest.NewRecorder()

	hf.ServeHTTP(w, r)
	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 12)
	Equal(t, test.Posted, "")
	Equal(t, test.MultiPartPosted, "value")

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)

	err = writer.WriteField("MultiPartPosted", "value")
	Equal(t, err, nil)

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	err = writer.Close()
	Equal(t, err, nil)

	test = new(TestStruct)
	r, _ = http.NewRequest(http.MethodPost, "/decode2/13", body)
	r.Header.Set(ContentType, writer.FormDataContentType())
	w = httptest.NewRecorder()

	hf.ServeHTTP(w, r)
	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 0)
	Equal(t, test.Posted, "")
	Equal(t, test.MultiPartPosted, "value")

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)

	err = writer.WriteField("MultiPartPosted", "value")
	Equal(t, err, nil)

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	err = writer.Close()
	Equal(t, err, nil)

	test = new(TestStruct)
	r, _ = http.NewRequest(http.MethodPost, "/decode3/11", body)
	r.Header.Set(ContentType, writer.FormDataContentType())
	w = httptest.NewRecorder()

	hf.ServeHTTP(w, r)
	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 11)
	Equal(t, test.Posted, "")
	Equal(t, test.MultiPartPosted, "value")

	jsonBody := `{"ID":13,"Posted":"value","MultiPartPosted":"value"}`
	test = new(TestStruct)
	r, _ = http.NewRequest(http.MethodPost, "/decode/13", strings.NewReader(jsonBody))
	r.Header.Set(ContentType, ApplicationJSON)
	w = httptest.NewRecorder()

	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 13)
	Equal(t, test.Posted, "value")
	Equal(t, test.MultiPartPosted, "value")

	xmlBody := `<TestStruct><ID>13</ID><Posted>value</Posted><MultiPartPosted>value</MultiPartPosted></TestStruct>`
	test = new(TestStruct)
	r, _ = http.NewRequest(http.MethodPost, "/decode/13", strings.NewReader(xmlBody))
	r.Header.Set(ContentType, ApplicationXML)
	w = httptest.NewRecorder()

	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 13)
	Equal(t, test.Posted, "value")
	Equal(t, test.MultiPartPosted, "value")
}

func TestAcceptedLanguages(t *testing.T) {

	req, _ := http.NewRequest("POST", "/", nil)
	req.Header.Set(AcceptedLanguage, "da, en-GB;q=0.8, en;q=0.7")

	languages := AcceptedLanguages(req)

	Equal(t, languages[0], "da")
	Equal(t, languages[1], "en-GB")
	Equal(t, languages[2], "en")

	req.Header.Del(AcceptedLanguage)

	languages = AcceptedLanguages(req)

	Equal(t, len(languages), 0)

	req.Header.Set(AcceptedLanguage, "")
	languages = AcceptedLanguages(req)

	Equal(t, len(languages), 0)
}

func TestAttachment(t *testing.T) {

	p := New()

	p.Get("/dl", func(w http.ResponseWriter, r *http.Request) {
		f, _ := os.Open("logo.png")
		if err := Attachment(w, f, "logo.png"); err != nil {
			panic(err)
		}
	})

	p.Get("/dl-unknown-type", func(w http.ResponseWriter, r *http.Request) {
		f, _ := os.Open("logo.png")
		if err := Attachment(w, f, "logo"); err != nil {
			panic(err)
		}
	})

	r, _ := http.NewRequest(http.MethodGet, "/dl", nil)
	w := &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
	hf := p.Serve()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentDisposition), "attachment;filename=logo.png")
	Equal(t, w.Header().Get(ContentType), "image/png")
	Equal(t, w.Body.Len(), 20797)

	r, _ = http.NewRequest(http.MethodGet, "/dl-unknown-type", nil)
	w = &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
	hf = p.Serve()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentDisposition), "attachment;filename=logo")
	Equal(t, w.Header().Get(ContentType), "application/octet-stream")
	Equal(t, w.Body.Len(), 20797)
}

func TestInline(t *testing.T) {

	p := New()
	p.Get("/dl-inline", func(w http.ResponseWriter, r *http.Request) {
		f, _ := os.Open("logo.png")
		if err := Inline(w, f, "logo.png"); err != nil {
			panic(err)
		}
	})

	p.Get("/dl-unknown-type-inline", func(w http.ResponseWriter, r *http.Request) {
		f, _ := os.Open("logo.png")
		if err := Inline(w, f, "logo"); err != nil {
			panic(err)
		}
	})

	r, _ := http.NewRequest(http.MethodGet, "/dl-inline", nil)
	w := &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
	hf := p.Serve()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentDisposition), "inline;filename=logo.png")
	Equal(t, w.Header().Get(ContentType), "image/png")
	Equal(t, w.Body.Len(), 20797)

	r, _ = http.NewRequest(http.MethodGet, "/dl-unknown-type-inline", nil)
	w = &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
	hf = p.Serve()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentDisposition), "inline;filename=logo")
	Equal(t, w.Header().Get(ContentType), "application/octet-stream")
	Equal(t, w.Body.Len(), 20797)
}

func TestClientIP(t *testing.T) {

	req, _ := http.NewRequest("POST", "/", nil)

	req.Header.Set("X-Real-IP", " 10.10.10.10  ")
	req.Header.Set("X-Forwarded-For", "  20.20.20.20, 30.30.30.30")
	req.RemoteAddr = "  40.40.40.40:42123 "

	Equal(t, ClientIP(req), "10.10.10.10")

	req.Header.Del("X-Real-IP")
	Equal(t, ClientIP(req), "20.20.20.20")

	req.Header.Set("X-Forwarded-For", "30.30.30.30  ")
	Equal(t, ClientIP(req), "30.30.30.30")

	req.Header.Del("X-Forwarded-For")
	Equal(t, ClientIP(req), "40.40.40.40")
}

func TestJSON(t *testing.T) {

	jsonData := `{"id":1,"name":"Patient Zero"}`
	callbackFunc := "CallbackFunc"

	p := New()
	p.Use(Gzip2)
	p.Get("/json", func(w http.ResponseWriter, r *http.Request) {
		if err := JSON(w, http.StatusOK, zombie{1, "Patient Zero"}); err != nil {
			panic(err)
		}
	})
	p.Get("/badjson", func(w http.ResponseWriter, r *http.Request) {
		if err := JSON(w, http.StatusOK, func() {}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	p.Get("/jsonbytes", func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal("Patient Zero")
		if err := JSONBytes(w, http.StatusOK, b); err != nil {
			panic(err)
		}
	})
	p.Get("/jsonp", func(w http.ResponseWriter, r *http.Request) {
		if err := JSONP(w, http.StatusOK, zombie{1, "Patient Zero"}, callbackFunc); err != nil {
			panic(err)
		}
	})
	p.Get("/badjsonp", func(w http.ResponseWriter, r *http.Request) {
		if err := JSONP(w, http.StatusOK, func() {}, callbackFunc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	hf := p.Serve()

	r, _ := http.NewRequest(http.MethodGet, "/json", nil)
	w := httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentType), ApplicationJSONCharsetUTF8)
	Equal(t, w.Body.String(), jsonData)

	r, _ = http.NewRequest(http.MethodGet, "/badjson", nil)
	w = httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusInternalServerError)
	Equal(t, w.Header().Get(ContentType), TextPlainCharsetUTF8)
	Equal(t, w.Body.String(), "json: unsupported type: func()\n")

	r, _ = http.NewRequest(http.MethodGet, "/jsonbytes", nil)
	w = httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentType), ApplicationJSONCharsetUTF8)
	Equal(t, w.Body.String(), "\"Patient Zero\"")

	r, _ = http.NewRequest(http.MethodGet, "/jsonp", nil)
	w = httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentType), ApplicationJavaScriptCharsetUTF8)
	Equal(t, w.Body.String(), callbackFunc+"("+jsonData+");")

	r, _ = http.NewRequest(http.MethodGet, "/badjsonp", nil)
	w = httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusInternalServerError)
	Equal(t, w.Header().Get(ContentType), TextPlainCharsetUTF8)
	Equal(t, w.Body.String(), "json: unsupported type: func()\n")
}

func TestXML(t *testing.T) {

	xmlData := `<zombie><id>1</id><name>Patient Zero</name></zombie>`

	p := New()
	p.Use(Gzip2)
	p.Get("/xml", func(w http.ResponseWriter, r *http.Request) {
		if err := XML(w, http.StatusOK, zombie{1, "Patient Zero"}); err != nil {
			panic(err)
		}
	})
	p.Get("/badxml", func(w http.ResponseWriter, r *http.Request) {
		if err := XML(w, http.StatusOK, func() {}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	p.Get("/xmlbytes", func(w http.ResponseWriter, r *http.Request) {
		b, _ := xml.Marshal(zombie{1, "Patient Zero"})
		if err := XMLBytes(w, http.StatusOK, b); err != nil {
			panic(err)
		}
	})

	hf := p.Serve()

	r, _ := http.NewRequest(http.MethodGet, "/xml", nil)
	w := httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentType), ApplicationXMLCharsetUTF8)
	Equal(t, w.Body.String(), xml.Header+xmlData)

	r, _ = http.NewRequest(http.MethodGet, "/xmlbytes", nil)
	w = httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentType), ApplicationXMLCharsetUTF8)
	Equal(t, w.Body.String(), xml.Header+xmlData)

	r, _ = http.NewRequest(http.MethodGet, "/badxml", nil)
	w = httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusInternalServerError)
	Equal(t, w.Header().Get(ContentType), TextPlainCharsetUTF8)
	Equal(t, w.Body.String(), "xml: unsupported type: func()\n")
}

func TestBadParseForm(t *testing.T) {

	// successful scenarios tested under TestDecode

	p := New()
	p.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		if err := ParseForm(r); err != nil {
			if _, errr := w.Write([]byte(err.Error())); errr != nil {
				panic(errr)
			}
			return
		}
	})

	code, body := request(http.MethodGet, "/users/16?test=%2f%%efg", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "invalid URL escape \"%%e\"")
}

func TestBadParseMultiPartForm(t *testing.T) {

	// successful scenarios tested under TestDecode

	p := New()
	p.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		if err := ParseMultipartForm(r, 10<<5); err != nil {
			if _, errr := w.Write([]byte(err.Error())); errr != nil {
				panic(err)
			}
			return
		}
	})

	code, body := requestMultiPart(http.MethodGet, "/users/16?test=%2f%%efg", p)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "invalid URL escape \"%%e\"")
}

type gzipWriter struct {
	io.Writer
	http.ResponseWriter
	sniffComplete bool
}

func (w *gzipWriter) Write(b []byte) (int, error) {

	if !w.sniffComplete {
		if w.Header().Get(ContentType) == "" {
			w.Header().Set(ContentType, http.DetectContentType(b))
		}
		w.sniffComplete = true
	}

	return w.Writer.Write(b)
}

func (w *gzipWriter) Flush() error {
	return w.Writer.(*gzip.Writer).Flush()
}

func (w *gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *gzipWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

var gzipPool = sync.Pool{
	New: func() interface{} {
		return &gzipWriter{Writer: gzip.NewWriter(ioutil.Discard)}
	},
}

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip2(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var gz *gzipWriter
		var gzr *gzip.Writer

		w.Header().Add(Vary, AcceptEncoding)

		if strings.Contains(r.Header.Get(AcceptEncoding), Gzip) {

			gz = gzipPool.Get().(*gzipWriter)
			gz.sniffComplete = false
			gzr = gz.Writer.(*gzip.Writer)
			gzr.Reset(w)
			gz.ResponseWriter = w

			w.Header().Set(ContentEncoding, Gzip)

			w = gz
			defer func() {

				// fmt.Println(gz.sniffComplete)
				if !gz.sniffComplete {
					// We have to reset response to it's pristine state when
					// nothing is written to body.
					w.Header().Del(ContentEncoding)
					gzr.Reset(ioutil.Discard)
				}

				gzr.Close()
				gzipPool.Put(gz)
			}()
		}

		next(w, r)
	}
}
