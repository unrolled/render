package render

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"net/http"
)

type Engine interface {
	Render(http.ResponseWriter, interface{}) error
}

type Head struct {
	ContentType string
	Status      int
}

//Built-in XML renderer
type XML struct {
	Head
	Indent bool
	Prefix []byte
}

//Built-in JSON renderer
type JSON struct {
	Head
	Indent bool
	Prefix []byte
}

//Build-in HTML renderer
type HTML struct {
	Head
	Name      string
	Templates *template.Template
}

//Built-in binary data renderer
type Data struct {
	Head
}

func (h Head) Write(w http.ResponseWriter) {
	w.Header().Set(ContentType, h.ContentType)
	w.WriteHeader(h.Status)
}

func (d Data) Render(w http.ResponseWriter, v interface{}) error {
	c := w.Header().Get(ContentType)
	if c != "" {
		d.Head.ContentType = c
	}

	d.Head.Write(w)
	w.Write(v.([]byte))
	return nil
}

func (j JSON) Render(w http.ResponseWriter, v interface{}) error {
	var result []byte
	var err error

	if j.Indent {
		result, err = json.MarshalIndent(v, "", "  ")
	} else {
		result, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}

	// JSON marshaled fine, write out the result.
	j.Head.Write(w)
	if len(j.Prefix) > 0 {
		w.Write(j.Prefix)
	}
	w.Write(result)
	return nil
}

func (x XML) Render(w http.ResponseWriter, v interface{}) error {
	var result []byte
	var err error

	if x.Indent {
		result, err = xml.MarshalIndent(v, "", "  ")
	} else {
		result, err = xml.Marshal(v)
	}
	if err != nil {
		return err
	}

	// XML marshaled fine, write out the result.
	x.Head.Write(w)
	if len(x.Prefix) > 0 {
		w.Write(x.Prefix)
	}
	w.Write(result)
	return nil
}

func (h HTML) Render(w http.ResponseWriter, binding interface{}) error {
	out := new(bytes.Buffer)
	err := h.Templates.ExecuteTemplate(out, h.Name, binding)
	if err != nil {
		return err
	}

	h.Head.Write(w)
	w.Write(out.Bytes())
	return nil
}
