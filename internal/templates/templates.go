// Package templates provides template rendering utilities.
package templates

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
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
			"linkifyURLs":               linkifyURLs,
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
	// Everyone is going to be in eastern time for now.
	tz, err := time.LoadLocation("America/New_York")
	if err != nil { // Always check errors even if they should not happen.
		panic(err)
	}
	return t.In(tz).Format("Jan 2, 2006 03:04:05 PM")
}

func fmtDate(t time.Time) string {
	// Everyone is going to be in eastern time for now.
	tz, err := time.LoadLocation("America/New_York")
	if err != nil { // Always check errors even if they should not happen.
		panic(err)
	}
	return t.In(tz).Format("Jan 2, 2006")
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

var (
	urlRegex       = regexp.MustCompile(`\bhttps?://[a-zA-Z0-9-\.]+(?::\d+)?(?:/[^\s]*)?\b`)
	formatLinkTmpl = template.Must(template.New("link").Parse(`<a href="{{.URL}}" rel="nofollow noreferrer">{{.URL}}</a>`))
)

type linkData struct {
	URL safehtml.URL
}

// linkifyURLs finds all URLs in the input and wraps them in <a> tags
// to make them hyperlinks when displayed in the browser.
func linkifyURLs(contents string) safehtml.HTML {
	matches := urlRegex.FindAllStringIndex(contents, -1)
	if len(matches) == 0 {
		return safehtml.HTMLEscaped(contents)
	}
	var chunks []safehtml.HTML
	prevEnd := 0
	for _, idx := range matches {
		start, end := idx[0], idx[1]
		if prevEnd < start {
			chunks = append(chunks, safehtml.HTMLEscaped(contents[prevEnd:start]))
		}
		d := linkData{
			URL: safehtml.URLSanitized(contents[start:end]),
		}
		wrapped, err := formatLinkTmpl.ExecuteTemplateToHTML("link", d)
		if err != nil {
			// Fallback, just concat the escaped URL to the contents.
			chunks = append(chunks, safehtml.HTMLEscaped(contents[start:end]))
		} else {
			chunks = append(chunks, wrapped)
		}
		prevEnd = end
	}
	if prevEnd < len(contents) {
		chunks = append(chunks, safehtml.HTMLEscaped(contents[prevEnd:]))
	}
	return safehtml.HTMLConcat(chunks...)
}
