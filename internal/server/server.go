// Package server provides the backend server for the site.
package server

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/google/safehtml/template"
	"github.com/jessesomerville/yodahunters/internal/templates"
)

// Config defines the backend server configuration.
type Config struct {
	Address    string
	TemplateFS template.TrustedFS
}

// Server handles HTTP connections and serves backend content.
type Server struct {
	renderer *templates.Renderer
}

// Run starts the server and returns an error upon exit.
func Run(cfg Config) error {
	renderer, err := templates.New(cfg.TemplateFS)
	if err != nil {
		return err
	}
	s := &Server{
		renderer: renderer,
	}

	mux := http.NewServeMux()
	mux.Handle("/", s.handleHome())
	log.Printf("Serving site at %q\n", cfg.Address)
	return http.ListenAndServe(cfg.Address, mux)
}

func (s *Server) serveHTML(w http.ResponseWriter, tmpl string, data any) {
	buf, err := s.renderer.Render(tmpl, data)
	if err != nil {
		log.Printf("serveHTML(w, %q, %+v): %v", tmpl, data, err)
		// TODO - Replace this with a proper error page response.
		http.Error(w, "failed to render page", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(w, bytes.NewReader(buf)); err != nil {
		log.Printf("Server.serveHTML(w, %q, data): failed to write to http.ResponseWriter: %v", tmpl, err)
	}
}

func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		s.serveHTML(w, "home", nil)
	}
}
