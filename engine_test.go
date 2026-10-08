package render

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type engineTestWriter struct {
	failAt  int
	partial bool
	writes  int
	body    bytes.Buffer
}

func (w *engineTestWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == w.failAt {
		n := 0
		if w.partial {
			n = len(p) / 2
		}
		_, _ = w.body.Write(p[:n])

		return n, io.ErrClosedPipe
	}

	return w.body.Write(p)
}

type engineTestResponseWriter struct {
	*engineTestWriter
	header       http.Header
	headerWrites int
	status       int
}

func (w *engineTestResponseWriter) Header() http.Header {
	return w.header
}

func (w *engineTestResponseWriter) WriteHeader(status int) {
	w.headerWrites++
	w.status = status
}

type engineWriteCase struct {
	name        string
	options     Options
	call        func(*Render, io.Writer, int) error
	parts       []string
	contentType string
}

func engineWriteCases() []engineWriteCase {
	data := func(r *Render, w io.Writer, status int) error {
		return r.Data(w, status, []byte("payload"))
	}
	text := func(r *Render, w io.Writer, status int) error {
		return r.Text(w, status, "payload")
	}
	html := func(r *Render, w io.Writer, status int) error {
		return r.HTML(w, status, "hello", "gophers")
	}
	json := func(r *Render, w io.Writer, status int) error {
		return r.JSON(w, status, map[string]string{"name": "value"})
	}
	jsonp := func(r *Render, w io.Writer, status int) error {
		return r.JSONP(w, status, "callback", map[string]string{"name": "value"})
	}
	xml := func(r *Render, w io.Writer, status int) error {
		return r.XML(w, status, GreetingXML{One: "hello", Two: "world"})
	}

	const (
		jsonBody         = `{"name":"value"}`
		indentedJSONBody = "{\n  \"name\": \"value\"\n}\n"
		xmlBody          = `<greeting one="hello" two="world"></greeting>`
		jsonPrefix       = ")]}'\n"
		xmlPrefix        = "<?xml version='1.0'?>\n"
		charset          = "; charset=UTF-8"
	)

	return []engineWriteCase{
		{
			name: "data", call: data, parts: []string{"payload"}, contentType: ContentBinary,
		},
		{
			name: "text", call: text, parts: []string{"payload"}, contentType: ContentText + charset,
		},
		{
			name: "html", call: html, parts: []string{"<h1>Hello gophers</h1>\n"}, contentType: ContentHTML + charset,
		},
		{
			name: "json", call: json, parts: []string{jsonBody}, contentType: ContentJSON + charset,
		},
		{
			name: "json-prefix", call: json, parts: []string{jsonPrefix, jsonBody}, contentType: ContentJSON + charset,
			options: Options{PrefixJSON: []byte(jsonPrefix)},
		},
		{
			name: "json-indented-prefix", call: json, parts: []string{jsonPrefix, indentedJSONBody}, contentType: ContentJSON + charset,
			options: Options{PrefixJSON: []byte(jsonPrefix), IndentJSON: true},
		},
		{
			name: "streaming-json", call: json, parts: []string{jsonBody + "\n"}, contentType: ContentJSON + charset,
			options: Options{StreamingJSON: true},
		},
		{
			name: "streaming-json-prefix", call: json, parts: []string{jsonPrefix, jsonBody + "\n"}, contentType: ContentJSON + charset,
			options: Options{PrefixJSON: []byte(jsonPrefix), StreamingJSON: true},
		},
		{
			name: "jsonp", call: jsonp, parts: []string{"callback(", jsonBody, ");"}, contentType: ContentJSONP + charset,
		},
		{
			name: "jsonp-indented", call: jsonp, parts: []string{"callback(", strings.TrimSuffix(indentedJSONBody, "\n"), ");", "\n"}, contentType: ContentJSONP + charset,
			options: Options{IndentJSON: true},
		},
		{
			name: "xml", call: xml, parts: []string{xmlBody}, contentType: ContentXML + charset,
		},
		{
			name: "xml-prefix", call: xml, parts: []string{xmlPrefix, xmlBody}, contentType: ContentXML + charset,
			options: Options{PrefixXML: []byte(xmlPrefix)},
		},
		{
			name: "xml-indented-prefix", call: xml, parts: []string{xmlPrefix, xmlBody + "\n"}, contentType: ContentXML + charset,
			options: Options{PrefixXML: []byte(xmlPrefix), IndentXML: true},
		},
	}
}

func TestRenderDestinationWriteErrors(t *testing.T) {
	for _, test := range engineWriteCases() {
		t.Run(test.name, func(t *testing.T) {
			options := test.options
			options.Directory = "testdata/basic"
			options.DisableHTTPErrorRendering = true
			render := New(options)

			for failAt := 1; failAt <= len(test.parts); failAt++ {
				for _, partial := range []bool{false, true} {
					for _, httpWriter := range []bool{false, true} {
						name := fmt.Sprintf("write%d/partial=%t/http=%t", failAt, partial, httpWriter)
						t.Run(name, func(t *testing.T) {
							writer := &engineTestWriter{failAt: failAt, partial: partial}
							var destination io.Writer = writer
							var response *engineTestResponseWriter
							if httpWriter {
								response = &engineTestResponseWriter{engineTestWriter: writer, header: make(http.Header)}
								destination = response
							}

							err := test.call(render, destination, http.StatusAccepted)
							expect(t, err, io.ErrClosedPipe)
							expect(t, writer.writes, failAt)

							want := strings.Join(test.parts[:failAt-1], "")
							if partial {
								failedPart := test.parts[failAt-1]
								want += failedPart[:len(failedPart)/2]
							}
							expect(t, writer.body.String(), want)
							if response != nil {
								expect(t, response.headerWrites, 1)
								expect(t, response.status, http.StatusAccepted)
								expect(t, response.Header().Get(ContentType), test.contentType)
							}
						})
					}
				}
			}
		})
	}
}

func TestRenderDestinationWriteSuccess(t *testing.T) {
	for _, test := range engineWriteCases() {
		t.Run(test.name, func(t *testing.T) {
			options := test.options
			options.Directory = "testdata/basic"
			options.DisableHTTPErrorRendering = true
			render := New(options)

			for _, httpWriter := range []bool{false, true} {
				t.Run(fmt.Sprintf("http=%t", httpWriter), func(t *testing.T) {
					writer := &engineTestWriter{}
					var destination io.Writer = writer
					var response *engineTestResponseWriter
					if httpWriter {
						response = &engineTestResponseWriter{engineTestWriter: writer, header: make(http.Header)}
						destination = response
					}

					expectNil(t, test.call(render, destination, http.StatusAccepted))
					expect(t, writer.body.String(), strings.Join(test.parts, ""))
					if response != nil {
						expect(t, response.headerWrites, 1)
						expect(t, response.status, http.StatusAccepted)
						expect(t, response.Header().Get(ContentType), test.contentType)
					}
				})
			}
		})
	}
}

func TestRenderEncodingErrorsBeforeDestinationWrite(t *testing.T) {
	cases := []struct {
		name string
		call func(*Render, io.Writer) error
	}{
		{
			name: "json",
			call: func(r *Render, w io.Writer) error { return r.JSON(w, http.StatusAccepted, math.NaN()) },
		},
		{
			name: "jsonp",
			call: func(r *Render, w io.Writer) error { return r.JSONP(w, http.StatusAccepted, "callback", math.NaN()) },
		},
		{
			name: "xml",
			call: func(r *Render, w io.Writer) error { return r.XML(w, http.StatusAccepted, func() {}) },
		},
		{
			name: "html",
			call: func(r *Render, w io.Writer) error { return r.HTML(w, http.StatusAccepted, "missing", nil) },
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			render := New(Options{
				Directory:                 "testdata/basic",
				DisableHTTPErrorRendering: true,
				PrefixJSON:                []byte("prefix"),
				PrefixXML:                 []byte("prefix"),
			})
			for _, httpWriter := range []bool{false, true} {
				t.Run(fmt.Sprintf("http=%t", httpWriter), func(t *testing.T) {
					writer := &engineTestWriter{failAt: 1}
					var destination io.Writer = writer
					var response *engineTestResponseWriter
					if httpWriter {
						response = &engineTestResponseWriter{engineTestWriter: writer, header: make(http.Header)}
						destination = response
					}

					expectNotNil(t, test.call(render, destination))
					expect(t, writer.writes, 0)
					expect(t, writer.body.Len(), 0)
					if response != nil {
						expect(t, response.headerWrites, 0)
						expect(t, response.Header().Get(ContentType), "")
					}
				})
			}
		})
	}
}

func TestRenderNoContentWriteError(t *testing.T) {
	for _, test := range engineWriteCases() {
		t.Run(test.name, func(t *testing.T) {
			options := test.options
			options.Directory = "testdata/basic"
			options.DisableHTTPErrorRendering = true
			render := New(options)
			writeErrors := make(chan error, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				writeErrors <- test.call(render, w, http.StatusNoContent)
			}))
			defer server.Close()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
			if err != nil {
				t.Fatal(err)
			}
			response, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				expectNil(t, response.Body.Close())
			}()

			body, err := io.ReadAll(response.Body)
			expectNil(t, err)
			expect(t, response.StatusCode, http.StatusNoContent)
			expect(t, string(body), "")
			expect(t, <-writeErrors, http.ErrBodyNotAllowed)
		})
	}
}
