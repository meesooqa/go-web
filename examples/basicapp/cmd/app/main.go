package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/meesooqa/go-cfg"
	"github.com/meesooqa/go-lgr"
	"github.com/meesooqa/go-middleware"
	"github.com/meesooqa/go-web"
)

// AppConfig general app config
type AppConfig struct {
	Logger lgr.Config `yaml:"logger"`
	Server web.Config `yaml:"server"`
	Auth   middleware.BasicAuthConfig
}

func main() {
	conf, err := cfg.Load[AppConfig]("etc/config.yml", ".env")
	if err != nil {
		log.Fatal(err)
	}
	logger, err := lgr.New(conf.Logger)
	if err != nil {
		log.Fatal(err)
	}
	slog.SetDefault(logger)

	srv, err := web.New(conf.Server, logger)
	if err != nil {
		logger.Error("fail to create server", "error", err)
		os.Exit(1)
	}

	srv.Use(
		middleware.Logging(logger),
		middleware.CORS(conf.Server.CORS),
	)

	home := &HomeController{
		tmpl:    srv.Templates(),
		authCfg: conf.Auth,
	}
	srv.Register(home)

	// Set custom 404 handler.
	srv.NotFoundHandler(home.notFound)

	// Graceful shutdown по Ctrl+C / SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("starting application")
	if err := srv.Run(ctx); err != nil {
		logger.Error("server exited with an error", "error", err)
		os.Exit(1)
	}
	logger.Info("application stopped")
}

type HomeController struct {
	tmpl    *web.Templates
	authCfg middleware.BasicAuthConfig
}

func (c *HomeController) Routes() []web.Route {
	auth := func(h http.HandlerFunc) http.HandlerFunc {
		handler := middleware.BasicAuth(c.authCfg)(h)
		return func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
		}
	}

	return []web.Route{
		{Pattern: "GET /{$}", Handler: c.index},
		{Pattern: "GET /about", Handler: c.about},
		{Pattern: "GET /admin", Handler: auth(c.admin)},
	}
}

func (c *HomeController) index(w http.ResponseWriter, r *http.Request) {
	if err := c.tmpl.Render(w, http.StatusOK, "index.html", map[string]any{
		"Title": "Main",
	}); err != nil {
		lgr.FromContext(r.Context()).Error("rendering error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (c *HomeController) about(w http.ResponseWriter, r *http.Request) {
	if err := c.tmpl.Render(w, http.StatusOK, "about.html", map[string]any{
		"Title": "About",
	}); err != nil {
		lgr.FromContext(r.Context()).Error("rendering error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (c *HomeController) admin(w http.ResponseWriter, r *http.Request) {
	if err := c.tmpl.Render(w, http.StatusOK, "admin.html", map[string]any{
		"Title": "Admin",
	}); err != nil {
		lgr.FromContext(r.Context()).Error("rendering error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (c *HomeController) notFound(w http.ResponseWriter, r *http.Request) {
	if err := c.tmpl.Render(w, http.StatusNotFound, "404.html", map[string]any{
		"Path": r.URL.Path,
	}); err != nil {
		lgr.FromContext(r.Context()).Error("rendering 404 error", "error", err)
		http.Error(w, "page not found", http.StatusNotFound)
	}
}
