// Package server provides the backend server for the site.
package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"

	"github.com/google/safehtml/template"
	"github.com/jessesomerville/yodahunters/internal/envconfig"
	"github.com/jessesomerville/yodahunters/internal/log"
	"github.com/jessesomerville/yodahunters/internal/pg"
	"github.com/jessesomerville/yodahunters/internal/server/middleware"
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

	jwtSecret []byte

	devmode bool
}

// Run starts the server and returns an error upon exit.
func Run(ctx context.Context, cfg Config) error {
	log.InitLogger()
	log.Infof(ctx, "Started logger")

	renderer, err := templates.New(cfg.TemplateFS)
	if err != nil {
		return err
	}
	dbname := envconfig.GetEnvOrDefault("YODAHUNTERS_DATABASE_NAME", "yodahunters-db")
	dbClient, err := pg.NewClient(ctx, dbname)
	if err != nil {
		return err
	}

	if err := pg.RunMigrations(ctx, dbClient); err != nil {
		return err
	}

	s := &Server{
		renderer: renderer,
		tmplFS:   cfg.TemplateFS,
		dbClient: dbClient,
		devmode:  cfg.DevMode,
	}

	s.jwtSecret = []byte(envconfig.GetEnvOrDefault("YODAHUNTERS_JWT_SECRET", ""))
	if len(s.jwtSecret) != 32 {
		log.Warnf(ctx, "Falling back to ephemeral JWT secret due to invalid YODAHUNTERS_JWT_SECRET")
		s.jwtSecret = make([]byte, 32)
		n, err := rand.Read(s.jwtSecret)
		if err != nil || n != 32 {
			return fmt.Errorf("failed to generate JWT signing key: %v", err)
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/", middleware.Chain(s.handleHome, s.jwtSecret))
	mux.Handle("GET /login", middleware.ErrorHandler(s.handleLogin))
	mux.Handle("GET /register", middleware.ErrorHandler(s.handleRegister))
	mux.Handle("GET /register/{regkey}", middleware.ErrorHandler(s.handleRegisterKey))
	mux.Handle("GET /new_thread", middleware.Chain(s.handleNewThread, s.jwtSecret))
	mux.Handle("GET /users/{id}", middleware.Chain(s.handleUsers, s.jwtSecret))
	mux.Handle("GET /users/edit", middleware.Chain(s.handleUsersEdit, s.jwtSecret))
	mux.Handle("GET /threads/{id}", middleware.Chain(s.handleThread, s.jwtSecret))
	mux.Handle("GET /category/{id}", middleware.Chain(s.handleCategory, s.jwtSecret))

	// TODO: Switch all the middleware to the full chain
	apiMux := http.NewServeMux()
	apiMux.Handle("GET /threads", middleware.Chain(s.apiHandleGetThreads, s.jwtSecret))
	apiMux.Handle("GET /category/{id}", middleware.Chain(s.getHandleGetThreadsByCategoryID, s.jwtSecret))
	apiMux.Handle("GET /threads/{id}", middleware.Chain(s.apiHandleGetThreadByID, s.jwtSecret))
	apiMux.Handle("GET /threads/{id}/comments", middleware.Chain(s.apiHandleGetCommentsByThreadID, s.jwtSecret))
	apiMux.Handle("POST /threads", middleware.Chain(s.apiHandlePostThreads, s.jwtSecret))
	// TODO?: delete

	apiMux.HandleFunc("POST /categories", middleware.AdminChain(s.apiHandlePostCategories, s.jwtSecret))
	apiMux.HandleFunc("GET /categories", middleware.Chain(s.apiHandleGetCategories, s.jwtSecret))

	apiMux.Handle("POST /comments", middleware.Chain(s.apiHandlePostComments, s.jwtSecret))
	apiMux.Handle("GET /comments/{id}", middleware.Chain(s.apiHandleGetCommentByID, s.jwtSecret))

	apiMux.HandleFunc("POST /register", middleware.ErrorHandler(s.apiHandleRegister))
	apiMux.HandleFunc("POST /login", middleware.ErrorHandler(s.apiHandleLogin))

	apiMux.HandleFunc("GET /me", middleware.Chain(s.apiHandleGetMe, s.jwtSecret))
	apiMux.HandleFunc("POST /me", middleware.Chain(s.apiHandlePostMe, s.jwtSecret))

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Infof(ctx, "Serving site at %q\n", cfg.Address)
	return http.ListenAndServe(cfg.Address, middleware.Logger(ctx, mux))
}

func (s *Server) serveHTML(ctx context.Context, w http.ResponseWriter, tmpl string, data any) error {
	if s.devmode {
		renderer, err := templates.New(s.tmplFS)
		if err != nil {
			return err
		}
		s.renderer = renderer
	}

	buf, err := s.renderer.Render(tmpl, data)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, bytes.NewReader(buf)); err != nil {
		log.Errorf(ctx, "serveHTML(w, %q, data): failed to write to http.ResponseWriter: %v", tmpl, err)
	}
	return nil
}
