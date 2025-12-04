// Package templates provides template rendering utilities.
package templates

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/google/safehtml/template"
)

// Renderer manages the parsing and rendering of template files.
type Renderer struct {
	tmpls sync.Map
}

// New returns a Renderer populated with the templates in the given filesystem.
func New(fs template.TrustedFS) (*Renderer, error) {
	pages := []string{"home", "login"}

	r := new(Renderer)
	for _, page := range pages {
		t, err := template.New("base.tmpl").ParseFS(fs, "*.tmpl")
		if err != nil {
			return nil, fmt.Errorf("ParseFS: %v", err)
		}
		p := filepath.Join(page, "*.tmpl")
		if _, err := t.ParseFS(fs, p); err != nil {
			return nil, fmt.Errorf("ParseFS(%q): %v", p, err)
		}
		r.tmpls.Store(page, t)
	}
	return r, nil
}

// Render renders the named template using the provided data.
func (r *Renderer) Render(name string, data any) ([]byte, error) {
	tmpl, ok := r.tmpls.Load(name)
	if !ok {
		return nil, fmt.Errorf("template named %q not found", name)
	}
	t := tmpl.(*template.Template)
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to render template for %q: %v", name, err)
	}
	return buf.Bytes(), nil
}
