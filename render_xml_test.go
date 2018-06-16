package render

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
)

type GreetingXML struct {
	XMLName xml.Name `xml:"greeting"`
	One     string   `xml:"one,attr"`
	Two     string   `xml:"two,attr"`
}

func TestXMLBasic(t *testing.T) {
	render := New(Options{
		// nothing here to configure
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.XML(w, 299, GreetingXML{One: "hello", Two: "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 299)
	expect(t, res.Header().Get(ContentType), ContentXML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<greeting one=\"hello\" two=\"world\"></greeting>")
}

func TestXMLPrefix(t *testing.T) {
	prefix := "<?xml version='1.0' encoding='UTF-8'?>\n"
	render := New(Options{
		PrefixXML: []byte(prefix),
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.XML(w, 300, GreetingXML{One: "hello", Two: "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentXML+"; charset=UTF-8")
	expect(t, res.Body.String(), prefix+"<greeting one=\"hello\" two=\"world\"></greeting>")
}

func TestXMLIndented(t *testing.T) {
	render := New(Options{
		IndentXML: true,
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.XML(w, http.StatusOK, GreetingXML{One: "hello", Two: "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get(ContentType), ContentXML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<greeting one=\"hello\" two=\"world\"></greeting>\n")
}

func TestXMLWithError(t *testing.T) {
	render := New(Options{
		// nothing here to configure
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.XML(w, 299, map[string]string{"foo": "bar"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNotNil(t, err)
	expect(t, res.Code, 500)
}

func TestXMLCustomContentType(t *testing.T) {
	render := New(Options{
		XMLContentType: "application/customxml",
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.XML(w, http.StatusOK, GreetingXML{One: "hello", Two: "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get(ContentType), "application/customxml; charset=UTF-8")
	expect(t, res.Body.String(), "<greeting one=\"hello\" two=\"world\"></greeting>")
}

func TestXMLDisabledCharset(t *testing.T) {
	render := New(Options{
		DisableCharset: true,
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.XML(w, http.StatusOK, GreetingXML{One: "hello", Two: "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get(ContentType), ContentXML)
	expect(t, res.Body.String(), "<greeting one=\"hello\" two=\"world\"></greeting>")
}
