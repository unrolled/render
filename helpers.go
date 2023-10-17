package render

import (
	"fmt"
	"html/template"
)

// Included helper functions for use when rendering HTML.
func helperFuncs() template.FuncMap {
	return template.FuncMap{
		"yield": func() (string, error) {
			return "", fmt.Errorf("yield called with no layout defined")
		},
		"partial": func() (string, error) {
			return "", fmt.Errorf("block called with no layout defined")
		},
		"current": func() (string, error) {
			return "", nil
		},
	}
}
