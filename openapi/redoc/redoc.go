package redoc

import (
	"bytes"
	"errors"
	"text/template"

	_ "embed"
)

// ErrSpecNotFound error for when spec file not found
var ErrSpecNotFound = errors.New("spec not found")

// Type configuration
type Type struct {
	SpecPath    string
	DocPath     string
	Title       string
	Description string
}

// HTML represents the redoc index.html page
//
//go:embed assets/index.html
var HTML string

// JavaScript represents the redoc standalone javascript
//
//go:embed assets/redoc.standalone.js
var JavaScript string

// Body returns the final html with the js in the body
func (r *Type) Body() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	tpl, err := template.New("redoc").Parse(HTML)
	if err != nil {
		return nil, err
	}
	if JavaScript == "" || HTML == "" {
		return nil, errors.New("redoc assets not found")
	}
	if err = tpl.Execute(buf, map[string]string{
		"body":        JavaScript,
		"title":       r.Title,
		"url":         r.SpecPath,
		"description": r.Description,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
