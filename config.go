package web

import "time"

// Config describes the web server configuration
type Config struct {
	// Host is the address the server listens on:
	// "0.0.0.0" for all interfaces,
	// "127.0.0.1" for local connections only.
	Host string `yaml:"host" default:"0.0.0.0"`

	// Port is the server's TCP port.
	Port int `yaml:"port" default:"8080" env:"SERVER_PORT"`

	// ReadTimeout is the maximum duration for reading the entire request,
	// including the body, from the time the connection is established.
	ReadTimeout time.Duration `yaml:"read_timeout" default:"15s"`

	// ReadHeaderTimeout is the maximum duration for reading request
	// headers. Provides separate protection against slow clients (Slowloris).
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" default:"5s"`

	// WriteTimeout is the maximum duration for writing the response.
	WriteTimeout time.Duration `yaml:"write_timeout" default:"15s"`

	// IdleTimeout is how long to keep a keep-alive connection open
	// between requests from the same client.
	IdleTimeout time.Duration `yaml:"idle_timeout" default:"60s"`

	// ShutdownTimeout is how long to wait for active requests to complete
	// during server shutdown (graceful shutdown) before forcibly
	// terminating them.
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" default:"10s"`

	// MaxHeaderBytes is the maximum total size of request headers
	// in bytes. Protects against excessively large headers.
	MaxHeaderBytes int `yaml:"max_header_bytes" default:"1048576"` // 1 MB

	// TemplatesDir is the directory with HTML templates (html/template),
	// scanned recursively. If empty, templates are not loaded
	// automatically; pass them via WithTemplates if needed.
	TemplatesDir string `yaml:"templates_dir"`

	// StaticDir is the directory with static files (CSS, JS, images).
	// If empty, static file serving is not enabled.
	StaticDir string `yaml:"static_dir"`

	// StaticURLPath is the URL prefix under which static files are served.
	StaticURLPath string `yaml:"static_url_path" default:"/static/"`
}
