// Package web provides a general-purpose web server for serving
// HTML pages (html/template)
package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/meesooqa/go-cfg"
	"github.com/meesooqa/go-middleware"
)

// Server is a web server: a router, a set of global middleware,
// (optionally) loaded templates, and a wrapper around http.Server
// with graceful shutdown support
type Server struct {
	config    Config
	mux       *http.ServeMux
	logger    *slog.Logger
	templates *Templates
	globalMW  []middleware.Middleware

	httpSrv *http.Server

	mu    sync.RWMutex
	addr  string
	ready chan struct{}
}

// Option configures a Server during creation via New
type Option func(*Server)

// WithTemplates passes pre-loaded templates instead of
// automatically loading them from Config.TemplatesDir
func WithTemplates(t *Templates) Option {
	return func(s *Server) { s.templates = t }
}

// New creates a Server from the provided Config.
// Use this function if Config is part of the application's overall configuration,
// already loaded separately (via cfg.Load[AppConfig] with a Server web.Config field).
// If logger is nil, slog.Default() is used
func New(c Config, logger *slog.Logger, opts ...Option) (*Server, error) {
	if logger == nil {
		logger = slog.Default()
	}

	s := &Server{
		config: c,
		mux:    http.NewServeMux(),
		logger: logger,
		ready:  make(chan struct{}),
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.templates == nil && c.TemplatesDir != "" {
		t, err := LoadTemplates(c.TemplatesDir)
		if err != nil {
			return nil, err
		}
		s.templates = t
	}

	if c.StaticDir != "" {
		s.Static(c.StaticURLPath, c.StaticDir)
	}

	return s, nil
}

// Load reads a YAML file (and optional .env files) via the cfg package
// and immediately builds a Server. Use this function if the application
// does not have a single unified config.yml and the server settings are the only
// thing that needs to be loaded
func Load(path string, logger *slog.Logger, envFiles ...string) (*Server, error) {
	c, err := cfg.Load[Config](path, envFiles...)
	if err != nil {
		return nil, err
	}
	return New(*c, logger)
}

// Templates returns the loaded templates (or nil if they were not
// loaded — TemplatesDir is empty and WithTemplates was not used)
func (s *Server) Templates() *Templates {
	return s.templates
}

// Use adds global middleware applied to all requests.
// Middleware are executed in the order they are added
// (the first added is the outermost, see middleware.Chain).
// Call before Run
func (s *Server) Use(mw ...middleware.Middleware) {
	s.globalMW = append(s.globalMW, mw...)
}

// Register registers the routes of one or more controllers
func (s *Server) Register(controllers ...Controller) {
	for _, c := range controllers {
		for _, route := range c.Routes() {
			s.mux.HandleFunc(route.Pattern, route.Handler)
		}
	}
}

// HandleFunc registers a single handler directly, bypassing
// controllers — for cases where a full Controller is overkill
func (s *Server) HandleFunc(pattern string, h http.HandlerFunc) {
	s.mux.HandleFunc(pattern, h)
}

// Handle registers an http.Handler directly
func (s *Server) Handle(pattern string, h http.Handler) {
	s.mux.Handle(pattern, h)
}

// Static enables serving static files from the dir directory under the
// urlPath URL prefix. It is called automatically from New if
// Config.StaticDir is set — use it directly only if you need
// to serve more than one static directory
func (s *Server) Static(urlPath, dir string) {
	//s.logger.Debug("registering static", "urlPath", urlPath, "dir", dir)
	if !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}
	if !strings.HasSuffix(urlPath, "/") {
		urlPath += "/"
	}
	fileServer := http.FileServer(http.Dir(dir))
	s.mux.Handle("GET "+urlPath, http.StripPrefix(urlPath, fileServer))
}

// Handler returns the final http.Handler — the router wrapped
// in the global middleware chain. It does not require starting a network listener,
// making it convenient for testing (httptest.NewServer(srv.Handler())) or if you want to
// manage the net.Listener and TLS yourself instead of calling Run
func (s *Server) Handler() http.Handler {
	return middleware.Chain(s.mux, s.globalMW...)
}

// Addr returns the address the server is actually listening on, after
// Run has started accepting connections (useful when Config.Port
// is 0 — the OS selects a free port). Until that point,
// it returns an empty string
func (s *Server) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.addr
}

// Run starts the server and blocks until ctx is canceled, after which
// it performs a graceful shutdown with the Config.ShutdownTimeout.
// Typical call from main:
//
//	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
//	defer stop()
//	if err := srv.Run(ctx); err != nil {
//		log.Fatal(err)
//	}
//
// Run can only be called once per Server instance
func (s *Server) Run(ctx context.Context) error {
	s.httpSrv = &http.Server{
		Handler:           s.Handler(),
		ReadTimeout:       s.config.ReadTimeout,
		ReadHeaderTimeout: s.config.ReadHeaderTimeout,
		WriteTimeout:      s.config.WriteTimeout,
		IdleTimeout:       s.config.IdleTimeout,
		MaxHeaderBytes:    s.config.MaxHeaderBytes,
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("web: failed to listen on %s: %w", addr, err)
	}

	s.mu.Lock()
	s.addr = ln.Addr().String()
	s.mu.Unlock()
	close(s.ready)

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("server started", "addr", ln.Addr().String())
		if err := s.httpSrv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err

	case <-ctx.Done():
		s.logger.Info("shutdown signal received, shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()

		if err := s.httpSrv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("web: error shutting down server: %w", err)
		}
		return nil
	}
}
