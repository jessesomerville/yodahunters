// Package server provides the backend server for the site.
package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/safehtml/template"
	"github.com/jessesomerville/yodahunters/internal/envconfig"
	"github.com/jessesomerville/yodahunters/internal/log"
	"github.com/jessesomerville/yodahunters/internal/pg"
	"github.com/jessesomerville/yodahunters/internal/server/middleware"
	"github.com/jessesomerville/yodahunters/internal/templates"
	"github.com/jessesomerville/yodahunters/static"
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

	store Store

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
	defer dbClient.Close(ctx)

	if err := pg.RunMigrations(ctx, dbClient); err != nil {
		return err
	}

	s := &Server{
		renderer: renderer,
		tmplFS:   cfg.TemplateFS,
		store:    newPGStore(dbClient),
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
	mux.Handle("/", s.chain(s.handleHome))
	mux.Handle("GET /login", middleware.ErrorHandler(s.handleLogin))
	mux.Handle("GET /register", middleware.ErrorHandler(s.handleRegister))
	mux.Handle("GET /register/{regkey}", middleware.ErrorHandler(s.handleRegisterKey))
	mux.Handle("GET /new_thread", s.chain(s.handleNewThread))
	mux.Handle("GET /users/{id}", s.chain(s.handleUsers))
	mux.Handle("GET /users/edit", s.chain(s.handleUsersEdit))
	mux.Handle("GET /threads/{id}", s.chain(s.handleThread))
	mux.Handle("GET /category/{id}", s.chain(s.handleCategory))

	// TODO: Switch all the middleware to the full chain
	apiMux := http.NewServeMux()
	apiMux.Handle("GET /threads", s.chain(s.apiHandleGetThreads))
	apiMux.Handle("GET /category/{id}", s.chain(s.getHandleGetThreadsByCategoryID))
	apiMux.Handle("GET /threads/{id}", s.chain(s.apiHandleGetThreadByID))
	apiMux.Handle("GET /threads/{id}/comments", s.chain(s.apiHandleGetCommentsByThreadID))
	apiMux.Handle("POST /threads", s.chain(s.apiHandlePostThreads))
	// TODO?: delete

	apiMux.HandleFunc("POST /categories", s.adminChain(s.apiHandlePostCategories))
	apiMux.HandleFunc("GET /categories", s.chain(s.apiHandleGetCategories))

	apiMux.Handle("POST /comments", s.chain(s.apiHandlePostComments))
	apiMux.Handle("GET /comments/{id}", s.chain(s.apiHandleGetCommentByID))

	apiMux.HandleFunc("POST /register", middleware.ErrorHandler(s.apiHandleRegister))
	apiMux.HandleFunc("POST /login", middleware.ErrorHandler(s.apiHandleLogin))

	apiMux.HandleFunc("GET /me", s.chain(s.apiHandleGetMe))
	apiMux.HandleFunc("POST /me", s.chain(s.apiHandlePostMe))

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServerFS(static.FS)))

	srv := &http.Server{Addr: cfg.Address, Handler: middleware.Logger(ctx, mux)}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
		case <-ctx.Done():
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Errorf(ctx, "HTTP server shutdown: %v", err)
		}
	}()

	log.Infof(ctx, "Serving site at %q\n", cfg.Address)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) chain(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return middleware.Chain(f, s.jwtSecret)
}

func (s *Server) adminChain(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return middleware.AdminChain(f, s.jwtSecret)
}

func (s *Server) serveHTML(ctx context.Context, w http.ResponseWriter, tmpl string, data any) error {
	renderer := s.renderer
	if s.devmode {
		var err error
		renderer, err = templates.New(s.tmplFS)
		if err != nil {
			return err
		}
	}

	buf, err := renderer.Render(tmpl, data)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, bytes.NewReader(buf)); err != nil {
		log.Errorf(ctx, "serveHTML(w, %q, data): failed to write to http.ResponseWriter: %v", tmpl, err)
	}
	return nil
}
