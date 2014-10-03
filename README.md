# Render [![GoDoc](https://godoc.org/github.com/unrolled/render?status.png)](http://godoc.org/github.com/unrolled/render)

Render is a package that provides functionality for easily rendering JSON, XML, and HTML templates. This package is based on the [Martini](https://github.com/go-martini/martini) [render](https://github.com/martini-contrib/render) work.

## Usage
Render can be used with pretty much any web framework providing you can access the `http.ResponseWriter` variable from your handler. The rendering functions simply wrap Go's existing functionality for marshaling and rendering the given data.

- HTML: Uses the [html/template](http://golang.org/pkg/html/template/) package to render HTML templates.
- JSON: Uses the [encoding/json](http://golang.org/pkg/encoding/json/) package to marshal data into a JSON-encoded response.
- XML: Uses the [encoding/xml](http://golang.org/pkg/encoding/xml/) package to marshal data into an XML-encoded response.
- Binary Data: Passes the incoming data straight through to the `http.ResponseWriter`.

~~~ go
// main.go
package main

import (
    "encoding/xml"
    "net/http"

    "gopkg.in/unrolled/render.v1"
)

type ExampleXml struct {
    XMLName xml.Name `xml:"example"`
    One     string   `xml:"one,attr"`
    Two     string   `xml:"two,attr"`
}

func main() {
    r := render.New(render.Options{})
    mux := http.NewServeMux()

    mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
        w.Write([]byte("Welcome, visit sub pages now."))
    })

    mux.HandleFunc("/data", func(w http.ResponseWriter, req *http.Request) {
        r.Data(w, http.StatusOK, []byte("Some binary data here."))
    })

    mux.HandleFunc("/json", func(w http.ResponseWriter, req *http.Request) {
        r.JSON(w, http.StatusOK, map[string]string{"hello": "json"})
    })

    mux.HandleFunc("/xml", func(w http.ResponseWriter, req *http.Request) {
        r.XML(w, http.StatusOK, ExampleXml{One: "hello", Two: "xml"})
    })

    mux.HandleFunc("/html", func(w http.ResponseWriter, req *http.Request) {
        // Assumes you have a template in ./templates called "example.tmpl"
        // $ mkdir -p templates && echo "<h1>Hello {{.}}.</h1>" > templates/example.tmpl
        r.HTML(w, http.StatusOK, "example", nil)
    })

    http.ListenAndServe("0.0.0.0:3000", mux)
}
~~~

~~~ html
<!-- templates/example.tmpl -->
<h1>Hello {{.}}.</h1>
~~~

### Options
`render.Render` comes with a variety of configuration options:

~~~ go
// ...
r := render.New(render.Options{
    Directory: "templates", // Specify what path to load the templates from.
    Layout: "layout", // Specify a layout template. Layouts can call {{ yield }} to render the current template.
    Extensions: []string{".tmpl", ".html"}, // Specify extensions to load for templates.
    Funcs: []template.FuncMap{AppHelpers}, // Specify helper function maps for templates to access.
    Delims: render.Delims{"{[{", "}]}"}, // Sets delimiters to the specified strings.
    Charset: "UTF-8", // Sets encoding for json and html content-types. Default is "UTF-8".
    IndentJSON: true, // Output human readable JSON.
    IndentXML: true, // Output human readable XML.
    PrefixJSON: []byte(")]}',\n"), // Prefixes JSON responses with the given bytes.
    PrefixXML: []byte("<?xml version='1.0' encoding='UTF-8'?>"), // Prefixes XML responses with the given bytes.
    HTMLContentType: "application/xhtml+xml", // Output XHTML content type instead of default "text/html".
    IsDevelopment: true, // Render will now recompile the templates on every HTML response.
})
// ...
~~~

### Loading Templates
By default the `render.Render` middleware will attempt to load templates with a '.tmpl' extension from the "templates" directory. Templates are found by traversing the templates directory and are named by path and basename. For instance, the following directory structure:

~~~
templates/
  |
  |__ admin/
  |      |
  |      |__ index.tmpl
  |      |
  |      |__ edit.tmpl
  |
  |__ home.tmpl
~~~

Will provide the following templates:
~~~
admin/index
admin/edit
home
~~~

### Layouts
`render.Render` provides a `yield` function for layouts to access:
~~~ go
// ...
r := render.New(render.Options{
    Layout: "layout",
})
// ...
~~~

~~~ html
<!-- templates/layout.tmpl -->
<html>
  <head>
    <title>My Layout</title>
  </head>
  <body>
    <!-- Render the current template here -->
    {{ yield }}
  </body>
</html>
~~~

`current` can also be called to get the current template being rendered.
~~~ html
<!-- templates/layout.tmpl -->
<html>
  <head>
    <title>My Layout</title>
  </head>
  <body>
    This is the {{ current }} page.
  </body>
</html>
~~~

### Character Encodings
The `render.Render` will automatically set the proper Content-Type header based on which function you call. See below for an example of what the default settings would output (note that UTF-8 is the default, and binary data does not output the charset):
~~~ go
// main.go
package main

import (
    "encoding/xml"
    "net/http"

    "gopkg.in/unrolled/render.v1"
)

type ExampleXml struct {
    XMLName xml.Name `xml:"example"`
    One     string   `xml:"one,attr"`
    Two     string   `xml:"two,attr"`
}

func main() {
    r := render.New(render.Options{})
    mux := http.NewServeMux()

    // This will set the Content-Type header to "application/octet-stream".
    // Note that this does not receive a charset value.
    mux.HandleFunc("/data", func(w http.ResponseWriter, req *http.Request) {
        r.Data(w, http.StatusOK, []byte("Some binary data here."))
    })

    // This will set the Content-Type header to "application/json; charset=UTF-8".
    mux.HandleFunc("/json", func(w http.ResponseWriter, req *http.Request) {
        r.JSON(w, http.StatusOK, map[string]string{"hello": "json"})
    })

    // This will set the Content-Type header to "text/xml; charset=UTF-8".
    mux.HandleFunc("/xml", func(w http.ResponseWriter, req *http.Request) {
        r.XML(w, http.StatusOK, ExampleXml{One: "hello", Two: "xml"})
    })

    // This will set the Content-Type header to "text/html; charset=UTF-8".
    mux.HandleFunc("/html", func(w http.ResponseWriter, req *http.Request) {
        // Assumes you have a template in ./templates called "example.tmpl"
        // $ mkdir -p templates && echo "<h1>Hello {{.}}.</h1>" > templates/example.tmpl
        r.HTML(w, http.StatusOK, "example", nil)
    })

    http.ListenAndServe("0.0.0.0:3000", mux)
}
~~~

In order to change the charset, you can set the `Charset` within the `render.Options` to your encoding value:
~~~ go
// main.go
package main

import (
    "encoding/xml"
    "net/http"

    "gopkg.in/unrolled/render.v1"
)

type ExampleXml struct {
    XMLName xml.Name `xml:"example"`
    One     string   `xml:"one,attr"`
    Two     string   `xml:"two,attr"`
}

func main() {
    r := render.New(render.Options{
        Charset: "ISO-8859-1",
    })
    mux := http.NewServeMux()

    // This will set the Content-Type header to "application/octet-stream".
    // Note that this does not receive a charset value.
    mux.HandleFunc("/data", func(w http.ResponseWriter, req *http.Request) {
        r.Data(w, http.StatusOK, []byte("Some binary data here."))
    })

    // This will set the Content-Type header to "application/json; charset=ISO-8859-1".
    mux.HandleFunc("/json", func(w http.ResponseWriter, req *http.Request) {
        r.JSON(w, http.StatusOK, map[string]string{"hello": "json"})
    })

    // This will set the Content-Type header to "text/xml; charset=ISO-8859-1".
    mux.HandleFunc("/xml", func(w http.ResponseWriter, req *http.Request) {
        r.XML(w, http.StatusOK, ExampleXml{One: "hello", Two: "xml"})
    })

    // This will set the Content-Type header to "text/html; charset=ISO-8859-1".
    mux.HandleFunc("/html", func(w http.ResponseWriter, req *http.Request) {
        // Assumes you have a template in ./templates called "example.tmpl"
        // $ mkdir -p templates && echo "<h1>Hello {{.}}.</h1>" > templates/example.tmpl
        r.HTML(w, http.StatusOK, "example", nil)
    })

    http.ListenAndServe("0.0.0.0:3000", mux)
}
~~~

## Integration Examples

### [Gin](https://github.com/gin-gonic/gin)
~~~ go
// main.go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "gopkg.in/unrolled/render.v1"
)

func main() {
    r := render.New(render.Options{
        IndentJSON: true,
    })

    masterGin := gin.Default()

    masterGin.GET("/", func(c *gin.Context) {
        r.JSON(c.Writer, http.StatusOK, map[string]string{"welcome": "This is rendered JSON!"})
    })

    masterGin.Run(":3000")
}
~~~

### [Goji](https://github.com/zenazn/goji)
~~~ go
// main.go
package main

import (
    "net/http"

    "github.com/zenazn/goji"
    "github.com/zenazn/goji/web"
    "gopkg.in/unrolled/render.v1"
)

func main() {
    r := render.New(render.Options{
        IndentJSON: true,
    })

    goji.Get("/", func(c web.C, w http.ResponseWriter, req *http.Request) {
        r.JSON(w, http.StatusOK, map[string]string{"welcome": "This is rendered JSON!"})
    })
    goji.Serve()  // Defaults to ":8000".
}
~~~

### [Negroni](https://github.com/codegangsta/negroni)
~~~ go
// main.go
package main

import (
    "net/http"

    "github.com/codegangsta/negroni"
    "gopkg.in/unrolled/render.v1"
)

func main() {
    r := render.New(render.Options{
        IndentJSON: true,
    })
    mux := http.NewServeMux()

    mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
        r.JSON(w, http.StatusOK, map[string]string{"welcome": "This is rendered JSON!"})
    })

    n := negroni.Classic()
    n.UseHandler(mux)
    n.Run(":3000")
}
~~~

### [Traffic](https://github.com/pilu/traffic/)
~~~ go
// main.go
package main

import (
    "net/http"

    "github.com/pilu/traffic"
    "gopkg.in/unrolled/render.v1"
)

func main() {
    r := render.New(render.Options{
        IndentJSON: true,
    })

    router := traffic.New()
    router.Get("/", func(w traffic.ResponseWriter, req *traffic.Request) {
        r.JSON(w, http.StatusOK, map[string]string{"welcome": "This is rendered JSON!"})
    })

    router.Run()
}
~~~

### [Web.go](https://github.com/hoisie/web)
~~~ go
// main.go
package main

import (
    "net/http"

    "github.com/hoisie/web"
    "gopkg.in/unrolled/render.v1"
)

func main() {
    r := render.New(render.Options{
        IndentJSON: true,
    })

    web.Get("/(.*)", func(ctx *web.Context, val string) {
        r.JSON(ctx, http.StatusOK, map[string]string{"welcome": "This is rendered JSON!"})
    })

    web.Run("0.0.0.0:3000")
}
~~~
