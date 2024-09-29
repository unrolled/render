package render

import (
	"errors"
	"html/template"
)

var (
	ErrYieldNoLayoutDefined = errors.New("yield called with no layout defined")
	ErrBlockNoLayoutDefined = errors.New("block called with no layout defined")
)

// Included helper functions for use when rendering HTML.
func helperFuncs() template.FuncMap {
	return template.FuncMap{
		"yield": func() (string, error) {
			return "", ErrYieldNoLayoutDefined
		},
		"partial": func() (string, error) {
			return "", ErrBlockNoLayoutDefined
		},
		"current": func() (string, error) {
			return "", nil
		},
	}
}
