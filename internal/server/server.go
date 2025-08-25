// Package server provides the backend server for the site.
package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/safehtml/template"
	"github.com/jessesomerville/yodahunters/internal/envconfig"
	"github.com/jessesomerville/yodahunters/internal/log"
	"github.com/jessesomerville/yodahunters/internal/pg"
	"github.com/jessesomerville/yodahunters/internal/templates"
)

// Config defines the backend server configuration.
type Config struct {
	// Address is the address to serve HTTP requests from.
	Address string
	// TemplateFS holds the golang templates for rendering HTML.
	TemplateFS template.TrustedFS
	// DevMode makes the server reparse the template files when a page
	// is loaded. This enables editing templates without having to restart
	// the server.
	DevMode bool
}

// Server handles HTTP connections and serves backend content.
type Server struct {
	renderer *templates.Renderer
	tmplFS   template.TrustedFS

	dbClient *pg.Client

	devmode bool
}

// Run starts the server and returns an error upon exit.
func Run(ctx context.Context, cfg Config) error {
	renderer, err := templates.New(cfg.TemplateFS)
	if err != nil {
		return err
	}
	dbname := envconfig.GetEnvOrDefault("YODAHUNTERS_DATABASE_NAME", "yodahunters-db")
	if err := pg.CreateDBIfNotExists(ctx, dbname); err != nil {
		return err
	}
	dbClient, err := pg.NewClient(ctx, dbname)
	if err != nil {
		return err
	}

	if cfg.DevMode {
		log.Warnf(ctx, "Dev mode is enabled, templates will be reparsed each time a page is loaded.")
	}

	s := &Server{
		renderer: renderer,
		tmplFS:   cfg.TemplateFS,
		dbClient: dbClient,
		devmode:  cfg.DevMode,
	}

	mux := http.NewServeMux()
	mux.Handle("/", s.handleHome())

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Infof(ctx, "Serving site at %q\n", cfg.Address)
	return http.ListenAndServe(cfg.Address, mux)
}

func (s *Server) serveHTML(ctx context.Context, w http.ResponseWriter, tmpl string, data any) {
	if s.devmode {
		renderer, err := templates.New(s.tmplFS)
		if err != nil {
			log.Errorf(ctx, "failed to reparse templates in devmode: %v", err)
			http.Error(w, fmt.Sprintf("[devmode]: failed to reparse templates: %v", err), http.StatusInternalServerError)
			return
		}
		s.renderer = renderer
	}

	buf, err := s.renderer.Render(tmpl, data)
	if err != nil {
		log.Errorf(ctx, "serveHTML(w, %q, data): %v", tmpl, err)
		// TODO - Replace this with a proper error page response.
		http.Error(w, "failed to render page", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(w, bytes.NewReader(buf)); err != nil {
		log.Errorf(ctx, "serveHTML(w, %q, data): failed to write to http.ResponseWriter: %v", tmpl, err)
	}
}

func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.serveHTML(r.Context(), w, "home", nil)
	}
}
