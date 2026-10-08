package render

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io"
	"net/http"
)

// Engine is the generic interface for all responses.
type Engine interface {
	Render(w io.Writer, v interface{}) error
}

// Head defines the basic ContentType and Status fields.
type Head struct {
	ContentType string
	Status      int
}

// Data built-in renderer.
type Data struct {
	Head
}

// HTML built-in renderer.
type HTML struct {
	Head
	Name      string
	Templates *template.Template

	bp GenericBufferPool
}

// JSONEncoder is the interface for encoding/json.Encoder.
type JSONEncoder interface {
	Encode(v interface{}) error
	SetEscapeHTML(on bool)
	SetIndent(prefix, indent string)
}

// JSON built-in renderer.
type JSON struct {
	Head
	Indent        bool
	UnEscapeHTML  bool
	Prefix        []byte
	StreamingJSON bool
	Encoder       func(w io.Writer) JSONEncoder
}

// JSONP built-in renderer.
type JSONP struct {
	Head
	Indent   bool
	Callback string
}

// Text built-in renderer.
type Text struct {
	Head
}

// XML built-in renderer.
type XML struct {
	Head
	Indent bool
	Prefix []byte
}

// Write outputs the header content.
func (h Head) Write(w http.ResponseWriter) {
	w.Header().Set(ContentType, h.ContentType)
	w.WriteHeader(h.Status)
}

// Render a data response.
func (d Data) Render(w io.Writer, v interface{}) error {
	if hw, ok := w.(http.ResponseWriter); ok {
		c := hw.Header().Get(ContentType)
		if c != "" {
			d.ContentType = c
		}

		d.Write(hw)
	}

	_, err := w.Write(v.([]byte))

	return err
}

// Render a HTML response.
func (h HTML) Render(w io.Writer, binding interface{}) error {
	var buf *bytes.Buffer
	if h.bp != nil {
		// If we have a bufferpool, allocate from it
		buf = h.bp.Get()
		defer h.bp.Put(buf)
	}

	err := h.Templates.ExecuteTemplate(buf, h.Name, binding)
	if err != nil {
		return err
	}

	if hw, ok := w.(http.ResponseWriter); ok {
		h.Write(hw)
	}

	_, err = buf.WriteTo(w)

	return err
}

// Render a JSON response.
func (j JSON) Render(w io.Writer, v interface{}) error {
	if j.StreamingJSON {
		return j.renderStreamingJSON(w, v)
	}

	var buf bytes.Buffer
	encoder := j.Encoder(&buf)
	encoder.SetEscapeHTML(!j.UnEscapeHTML)

	if j.Indent {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(v); err != nil {
		return err
	}

	output := buf.Bytes()

	// JSON marshaled fine, write out the result.
	if hw, ok := w.(http.ResponseWriter); ok {
		j.Write(hw)
	}

	if len(j.Prefix) > 0 {
		if _, err := w.Write(j.Prefix); err != nil {
			return err
		}
	}

	// Remove the newline that json.Encode injects when not indenting the output.
	if !j.Indent {
		output = bytes.TrimSuffix(output, []byte("\n"))
	}

	_, err := w.Write(output)

	return err
}

func (j JSON) renderStreamingJSON(w io.Writer, v interface{}) error {
	if hw, ok := w.(http.ResponseWriter); ok {
		j.Write(hw)
	}

	if len(j.Prefix) > 0 {
		if _, err := w.Write(j.Prefix); err != nil {
			return err
		}
	}

	encoder := j.Encoder(w)
	encoder.SetEscapeHTML(!j.UnEscapeHTML)

	if j.Indent {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(v)
}

// Render a JSONP response.
func (j JSONP) Render(w io.Writer, v interface{}) error {
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
	if hw, ok := w.(http.ResponseWriter); ok {
		j.Write(hw)
	}

	if _, err = w.Write([]byte(j.Callback + "(")); err != nil {
		return err
	}
	if _, err = w.Write(result); err != nil {
		return err
	}
	if _, err = w.Write([]byte(");")); err != nil {
		return err
	}

	// If indenting, append a new line.
	if j.Indent {
		_, err = w.Write([]byte("\n"))
	}

	return err
}

// Render a text response.
func (t Text) Render(w io.Writer, v interface{}) error {
	if hw, ok := w.(http.ResponseWriter); ok {
		c := hw.Header().Get(ContentType)
		if c != "" {
			t.ContentType = c
		}

		t.Write(hw)
	}

	_, err := w.Write([]byte(v.(string)))

	return err
}

// Render an XML response.
func (x XML) Render(w io.Writer, v interface{}) error {
	var result []byte

	var err error

	if x.Indent {
		result, err = xml.MarshalIndent(v, "", "  ")
		result = append(result, '\n')
	} else {
		result, err = xml.Marshal(v)
	}

	if err != nil {
		return err
	}

	// XML marshaled fine, write out the result.
	if hw, ok := w.(http.ResponseWriter); ok {
		x.Write(hw)
	}

	if len(x.Prefix) > 0 {
		if _, err = w.Write(x.Prefix); err != nil {
			return err
		}
	}

	_, err = w.Write(result)

	return err
}
