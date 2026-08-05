package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// homeController
type homeController struct {
	tmpl *Templates
}

func (c *homeController) Routes() []Route {
	return []Route{
		{Pattern: "GET /", Handler: c.index},
		{Pattern: "GET /ping", Handler: c.ping},
	}
}

func (c *homeController) index(w http.ResponseWriter, r *http.Request) {
	_ = c.tmpl.Render(w, http.StatusOK, "index.html", map[string]any{
		"Title": "Main", "Message": "test",
	})
}

func (c *homeController) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func testConfig() Config {
	return Config{
		Host:              "127.0.0.1",
		Port:              0,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		ShutdownTimeout:   0,
		MaxHeaderBytes:    1 << 20,
		StaticURLPath:     "/static/",
	}
}

func TestNew_WithoutTemplatesDir(t *testing.T) {
	srv, err := New(testConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv.Templates() != nil {
		t.Error("Templates() should be nil if TemplatesDir is not set")
	}
}

func TestNew_WithTemplatesDirOption(t *testing.T) {
	tmpl, err := LoadTemplates("testdata/templates")
	if err != nil {
		t.Fatalf("unexpected error loading templates: %v", err)
	}

	srv, err := New(testConfig(), nil, WithTemplates(tmpl))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv.Templates() == nil {
		t.Error("Templates() should not be nil after WithTemplates")
	}
}

func TestNew_TemplatesDirFromConfig_MissingDirFails(t *testing.T) {
	c := testConfig()
	c.TemplatesDir = "testdata/does_not_exist"

	_, err := New(c, nil)
	if err == nil {
		t.Fatal("expected an error for a non-existent TemplatesDir from Config")
	}
}

func TestServer_RegisterAndServe(t *testing.T) {
	tmpl, err := LoadTemplates("testdata/templates")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv, err := New(testConfig(), nil, WithTemplates(tmpl))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv.Register(&homeController{tmpl: tmpl})

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/ping")
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, expected 200", resp.StatusCode)
	}
}

func TestServer_Static(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "style.css"), []byte("body{color:red}"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	srv, err := New(testConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	srv.Static("/static", dir)

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/static/style.css")
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, expected 200", resp.StatusCode)
	}
}

func TestServer_StaticFromConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("console.log(1)"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	c := testConfig()
	c.StaticDir = dir
	c.StaticURLPath = "/assets/"

	srv, err := New(c, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/assets/app.js")
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, expected 200 (StaticDir from Config should be enabled automatically)", resp.StatusCode)
	}
}

func TestServer_Use_AppliesGlobalMiddleware(t *testing.T) {
	srv, err := New(testConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Global-Middleware", "applied")
			next.ServeHTTP(w, r)
		})
	})
	srv.HandleFunc("GET /check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/check")
	if err != nil {
		t.Fatalf("request expected: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("X-Global-Middleware"); got != "applied" {
		t.Errorf("X-Global-Middleware = %q, expected %q", got, "applied")
	}
}
