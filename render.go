// Render is a package that provides functionality for easily rendering JSON, XML, and HTML templates.
//
// package main
//
// import (
//   "encoding/xml"
//   "net/http"
//
//   "github.com/unrolled/render"
// )
//
// type ExampleXml struct {
//   XMLName xml.Name `xml:"example"`
//   One     string   `xml:"one,attr"`
//   Two     string   `xml:"two,attr"`
// }
//
// func main() {
//   r := render.New(render.Options{})
//   mux := http.NewServeMux()
//
//   mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
//     w.Write([]byte("Welcome, visit sub pages now."))
//   })
//
//   mux.HandleFunc("/data", func(w http.ResponseWriter, req *http.Request) {
//     r.Data(w, http.StatusOK, []byte("Some binary data here."))
//   })
//
//   mux.HandleFunc("/json", func(w http.ResponseWriter, req *http.Request) {
//     r.JSON(w, http.StatusOK, map[string]string{"hello": "json"})
//   })
//
//   mux.HandleFunc("/xml", func(w http.ResponseWriter, req *http.Request) {
//     r.XML(w, http.StatusOK, ExampleXml{One: "hello", Two: "xml"})
//   })
//
//   mux.HandleFunc("/html", func(w http.ResponseWriter, req *http.Request) {
//     // Assumes you have a template in ./templates called "example.tmpl"
//     // $ mkdir -p templates && echo "<h1>Hello HTML world.</h1>" > templates/example.tmpl
//     r.HTML(w, http.StatusOK, "example", nil)
//   })
//
//  http.ListenAndServe("0.0.0.0:3000", mux)
// }

package render

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	ContentType    = "Content-Type"
	ContentLength  = "Content-Length"
	ContentBinary  = "application/octet-stream"
	ContentJSON    = "application/json"
	ContentHTML    = "text/html"
	ContentXHTML   = "application/xhtml+xml"
	ContentXML     = "text/xml"
	defaultCharset = "UTF-8"
)

// Included helper functions for use when rendering html.
var helperFuncs = template.FuncMap{
	"yield": func() (string, error) {
		return "", fmt.Errorf("yield called with no layout defined")
	},
	"current": func() (string, error) {
		return "", nil
	},
}

// Delims represents a set of Left and Right delimiters for HTML template rendering.
type Delims struct {
	// Left delimiter, defaults to {{.
	Left string
	// Right delimiter, defaults to }}.
	Right string
}

// Options is a struct for specifying configuration options for the render.Render object.
type Options struct {
	// Directory to load templates. Default is "templates".
	Directory string
	// Layout template name. Will not render a layout if blank (""). Defaults to blank ("").
	Layout string
	// Extensions to parse template files from. Defaults to [".tmpl"].
	Extensions []string
	// Funcs is a slice of FuncMaps to apply to the template upon compilation. This is useful for helper functions. Defaults to [].
	Funcs []template.FuncMap
	// Delims sets the action delimiters to the specified strings in the Delims struct.
	Delims Delims
	// Appends the given character set to the Content-Type header. Default is "UTF-8".
	Charset string
	// Outputs human readable JSON.
	IndentJSON bool
	// Outputs human readable XML.
	IndentXML bool
	// Prefixes the JSON output with the given bytes.
	PrefixJSON []byte
	// Prefixes the XML output with the given bytes.
	PrefixXML []byte
	// Allows changing of output to XHTML instead of HTML. Default is "text/html"
	HTMLContentType string
	// If IsDevelopment is set to true, this will recompile the templates on every request. Default if false.
	IsDevelopment bool
}

// HTMLOptions is a struct for overriding some rendering Options for specific HTML call.
type HTMLOptions struct {
	// Layout template name. Overrides Options.Layout.
	Layout string
}

// Render is a service that provides functions for easily writing JSON, XML,
// Binary Data, and HTML templates out to a http Response.
type Render struct {
	// Customize Secure with an Options struct.
	opt             Options
	templates       *template.Template
	compiledCharset string
}

// Constructs a new Render instance with the supplied options.
func New(options Options) *Render {
	r := Render{
		opt: options,
	}

	r.prepareOptions()
	r.compileTemplates()

	return &r
}

func (r *Render) prepareOptions() {
	// Fill in the defaults if need be.
	if len(r.opt.Charset) == 0 {
		r.opt.Charset = defaultCharset
	}
	r.compiledCharset = "; charset=" + r.opt.Charset

	if len(r.opt.Directory) == 0 {
		r.opt.Directory = "templates"
	}
	if len(r.opt.Extensions) == 0 {
		r.opt.Extensions = []string{".tmpl"}
	}
	if len(r.opt.HTMLContentType) == 0 {
		r.opt.HTMLContentType = ContentHTML
	}
}

func (r *Render) compileTemplates() {
	dir := r.opt.Directory
	r.templates = template.New(dir)
	r.templates.Delims(r.opt.Delims.Left, r.opt.Delims.Right)

	// Walk the supplied directory and compile any files that match our extension list.
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		ext := ""
		if strings.Index(rel, ".") != -1 {
			ext = "." + strings.Join(strings.Split(rel, ".")[1:], ".")
		}

		for _, extension := range r.opt.Extensions {
			if ext == extension {

				buf, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}

				name := (rel[0 : len(rel)-len(ext)])
				tmpl := r.templates.New(filepath.ToSlash(name))

				// Add our funcmaps.
				for _, funcs := range r.opt.Funcs {
					tmpl.Funcs(funcs)
				}

				// Break out if this parsing fails. We don't want any silent server starts.
				template.Must(tmpl.Funcs(helperFuncs).Parse(string(buf)))
				break
			}
		}

		return nil
	})
}

// Marshals the given interface object and writes the JSON response.
func (r *Render) JSON(w http.ResponseWriter, status int, v interface{}) {
	var result []byte
	var err error
	if r.opt.IndentJSON {
		result, err = json.MarshalIndent(v, "", "  ")
	} else {
		result, err = json.Marshal(v)
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// JSON marshaled fine, write out the result.
	w.Header().Set(ContentType, ContentJSON+r.compiledCharset)
	w.WriteHeader(status)
	if len(r.opt.PrefixJSON) > 0 {
		w.Write(r.opt.PrefixJSON)
	}
	w.Write(result)
}

// Marshals the given interface object and writes the XML response.
func (r *Render) XML(w http.ResponseWriter, status int, v interface{}) {
	var result []byte
	var err error
	if r.opt.IndentXML {
		result, err = xml.MarshalIndent(v, "", "  ")
	} else {
		result, err = xml.Marshal(v)
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// XML marshaled fine, write out the result.
	w.Header().Set(ContentType, ContentXML+r.compiledCharset)
	w.WriteHeader(status)
	if len(r.opt.PrefixXML) > 0 {
		w.Write(r.opt.PrefixXML)
	}
	w.Write(result)
}

// Writes out the raw bytes as binary data.
func (r *Render) Data(w http.ResponseWriter, status int, v []byte) {
	if w.Header().Get(ContentType) == "" {
		w.Header().Set(ContentType, ContentBinary)
	}
	w.WriteHeader(status)
	w.Write(v)
}

// Builds up the HTML response from the specified template and bindings.
func (r *Render) HTML(w http.ResponseWriter, status int, name string, binding interface{}, htmlOpt ...HTMLOptions) {
	// If we are in development mode, recompile the templates on every HTML request.
	if r.opt.IsDevelopment {
		r.compileTemplates()
	}

	opt := r.prepareHTMLOptions(htmlOpt)

	// Assign a layout if there is one.
	if len(opt.Layout) > 0 {
		r.addYield(name, binding)
		name = opt.Layout
	}

	out, err := r.execute(name, binding)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Template rendered fine, write out the result.
	w.Header().Set(ContentType, r.opt.HTMLContentType+r.compiledCharset)
	w.WriteHeader(status)
	w.Write(out.Bytes())
}

func (r *Render) execute(name string, binding interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	return buf, r.templates.ExecuteTemplate(buf, name, binding)
}

func (r *Render) addYield(name string, binding interface{}) {
	funcs := template.FuncMap{
		"yield": func() (template.HTML, error) {
			buf, err := r.execute(name, binding)
			// Return safe HTML here since we are rendering our own template.
			return template.HTML(buf.String()), err
		},
		"current": func() (string, error) {
			return name, nil
		},
	}
	r.templates.Funcs(funcs)
}

func (r *Render) prepareHTMLOptions(htmlOpt []HTMLOptions) HTMLOptions {
	if len(htmlOpt) > 0 {
		return htmlOpt[0]
	}

	return HTMLOptions{
		Layout: r.opt.Layout,
	}
}
