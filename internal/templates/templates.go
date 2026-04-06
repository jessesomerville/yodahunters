// Package templates provides template rendering utilities.
package templates

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
	"github.com/google/safehtml/uncheckedconversions"
	"rsc.io/markdown"
)

// Renderer manages the parsing and rendering of template files.
type Renderer struct {
	tmpls sync.Map
}

// New returns a Renderer populated with the templates in the given filesystem.
func New(fs template.TrustedFS) (*Renderer, error) {
	pages := []string{"home", "login", "new_thread", "users", "edit_profile", "thread", "category", "register", "register_key"}

	r := new(Renderer)
	for _, page := range pages {
		t, err := template.New("base.tmpl").Funcs(template.FuncMap{
			// Registering a template function to convert timestamps to formatted strings
			"fmtTime":                   fmtTime,
			"fmtDate":                   fmtDate,
			"generateCommentID":         generateCommentID,
			"generateLatestCommentLink": generateLatestCommentLink,
			"renderMarkdown":            renderMarkdown,
		}).ParseFS(fs, "*.tmpl")
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

func fmtTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func fmtDate(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func generateCommentID(i int) safehtml.Identifier {
	return safehtml.IdentifierFromConstantPrefix("comment", strconv.Itoa(i))
}

func generateLatestCommentLink(threadID, replyCount, commentID int) string {
	page := replyCount/20 + 1
	if commentID == 0 {
		return fmt.Sprintf("/threads/%d", threadID)
	}
	return fmt.Sprintf("/threads/%d?page_number=%d#comment-%d", threadID, page, commentID)
}

var parser = markdown.Parser{
	AutoLinkText: true,
	Table:        true,
}

// renderMarkdown renders any markdown present in the input as HTML.
func renderMarkdown(contents string) safehtml.HTML {
	sanitized := safehtml.HTMLEscaped(contents)
	doc := parser.Parse(sanitized.String())
	return uncheckedconversions.HTMLFromStringKnownToSatisfyTypeContract(markdown.ToHTML(doc))
}
