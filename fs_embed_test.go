// +build go1.16

package render

import (
	"context"
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"
)

//go:embed testdata/*/*.html testdata/*/*.tmpl testdata/*/*/*.tmpl
var EmbedFixtures embed.FS

func TestEmbedFileSystemTemplateLookup(t *testing.T) {
	baseDir := "testdata/template-dir-test"
	fname0Rel := "0"
	fname1Rel := "subdir/1"
	fnameShouldParsedRel := "dedicated.tmpl/notbad"
	dirShouldNotParsedRel := "dedicated"

	r := New(Options{
		Directory:  baseDir,
		Extensions: []string{".tmpl", ".html"},
		FileSystem: &EmbedFileSystem{
			FS: EmbedFixtures,
		},
	})

	expect(t, r.TemplateLookup(fname1Rel) != nil, true)
	expect(t, r.TemplateLookup(fname0Rel) != nil, true)
	expect(t, r.TemplateLookup(fnameShouldParsedRel) != nil, true)
	expect(t, r.TemplateLookup(dirShouldNotParsedRel) == nil, true)
}

func TestEmbedFileSystemHTMLBasic(t *testing.T) {
	render := New(Options{
		Directory: "testdata/basic",
		FileSystem: &EmbedFileSystem{
			FS: EmbedFixtures,
		},
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "hello", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Hello gophers</h1>\n")
}
