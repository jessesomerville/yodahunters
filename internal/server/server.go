// Package server provides the backend server for the site.
package server

import (
	"bytes"
	"context"
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
	Address    string
	TemplateFS template.TrustedFS
}

// Server handles HTTP connections and serves backend content.
type Server struct {
	renderer *templates.Renderer
	dbClient *pg.Client
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

	s := &Server{
		renderer: renderer,
		dbClient: dbClient,
	}

	mux := http.NewServeMux()
	mux.Handle("/", s.handleHome())
	log.Infof(ctx, "Serving site at %q\n", cfg.Address)
	return http.ListenAndServe(cfg.Address, mux)
}

func (s *Server) serveHTML(ctx context.Context, w http.ResponseWriter, tmpl string, data any) {
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
