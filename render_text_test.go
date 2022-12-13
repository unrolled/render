package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTextBasic(t *testing.T) {
	render := New(Options{
		// nothing here to configure
	})

	var err error

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.Text(w, 299, "Hello Text!")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 299)
	expect(t, res.Header().Get(ContentType), ContentText+"; charset=UTF-8")
	expect(t, res.Body.String(), "Hello Text!")
}

func TestTextCharset(t *testing.T) {
	render := New(Options{
		Charset: "foobar",
	})

	var err error

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.Text(w, 299, "Hello Text!")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 299)
	expect(t, res.Header().Get(ContentType), ContentText+"; charset=foobar")
	expect(t, res.Body.String(), "Hello Text!")
}

func TestTextSuppliedCharset(t *testing.T) {
	render := New(Options{
		Charset: "foobar",
	})

	var err error

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContentType, "text/css")
		err = render.Text(w, 200, "html{color:red}")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), "text/css")
	expect(t, res.Body.String(), "html{color:red}")
}

func TestTextCustomContentType(t *testing.T) {
	render := New(Options{
		TextContentType: "application/customtext",
	})

	var err error

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.Text(w, http.StatusOK, "Hello Text!")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get(ContentType), "application/customtext; charset=UTF-8")
	expect(t, res.Body.String(), "Hello Text!")
}

func TestTextDisabledCharset(t *testing.T) {
	render := New(Options{
		DisableCharset: true,
	})

	var err error

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.Text(w, http.StatusOK, "Hello Text!")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get(ContentType), ContentText)
	expect(t, res.Body.String(), "Hello Text!")
}
