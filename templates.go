package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

// Templates is a wrapper around a set of HTML templates
// loaded from a directory. It uses the html/template package
// (instead of text/template) — values are automatically escaped
// when inserted into HTML, protecting against XSS vulnerabilities when rendering
// user data.
type Templates struct {
	tmpl *template.Template
}

type templatesOptions struct {
	extensions []string
	funcs      template.FuncMap
}

// TemplatesOption configures LoadTemplates.
type TemplatesOption func(*templatesOptions)

// WithExtensions specifies the file extensions that are considered
// templates (by default, only ".html").
func WithExtensions(exts ...string) TemplatesOption {
	return func(o *templatesOptions) { o.extensions = exts }
}

// WithFuncs adds custom functions available within
// templates (similar to template.FuncMap).
func WithFuncs(fm template.FuncMap) TemplatesOption {
	return func(o *templatesOptions) { o.funcs = fm }
}

// LoadTemplates recursively scans dir and parses all found
// template files using html/template. The name of each template within
// the set is its filename (e.g., "index.html"), so layouts
// and partials can be organized into subdirectories arbitrarily — the only requirement is
// that filenames must be unique within dir.
func LoadTemplates(dir string, opts ...TemplatesOption) (*Templates, error) {
	o := templatesOptions{extensions: []string{".html"}}
	for _, opt := range opts {
		opt(&o)
	}

	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		for _, ext := range o.extensions {
			if strings.HasSuffix(path, ext) {
				files = append(files, path)
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("web: failed to scan template directory %q: %w", dir, err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("web: no template files found in directory %q (extensions: %v)", dir, o.extensions)
	}

	t := template.New(filepath.Base(dir))
	if o.funcs != nil {
		t = t.Funcs(o.funcs)
	}

	t, err = t.ParseFiles(files...)
	if err != nil {
		return nil, fmt.Errorf("web: failed to parse templates from %q: %w", dir, err)
	}

	t = t.Option("missingkey=error")

	return &Templates{tmpl: t}, nil
}

// Render renders the template name with the given data and writes the result to w
// with the specified HTTP status. The template is first rendered into a temporary
// buffer — if rendering fails (e.g., due to a missing field in data), the client will not receive a half-sent
// page mixed with a 200 status code: headers and body are written to the response
// only when rendering is guaranteed to succeed.
func (t *Templates) Render(w http.ResponseWriter, status int, name string, data any) error {
	var buf bytes.Buffer
	if err := t.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return fmt.Errorf("web: failed to render template %q: %w", name, err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err := buf.WriteTo(w)
	return err
}
