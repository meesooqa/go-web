# go-web

`go-web` is a lightweight, general-purpose web server wrapper for Go, specifically designed for applications serving HTML pages using `html/template`. It provides a structured way to organize routes via controllers, handles template management, and ensures a robust server lifecycle with graceful shutdown.

## Key Features

- **Controller-Based Routing**: Organize related routes into controllers for better maintainability.
- **Automatic Template Loading**: Recursively loads all `.html` files from a specified directory.
- **Safe Rendering**: Templates are rendered to a buffer before being sent to the client, preventing "half-sent" pages with a 200 OK status on error.
- **Graceful Shutdown**: Built-in support for shutting down the server cleanly when a signal is received.
- **Flexible Configuration**: Supports YAML configuration and environment variables via the `go-cfg` package.
- **Static File Serving**: Easy setup for serving CSS, JS, and images.
- **Modern Go Standards**: Leverages `net/http` (Go 1.22+ routing) and `log/slog`.

## Installation

```bash
go get github.com/meesooqa/go-web
```

## Quick Start

### 1. Define a Controller

Create a struct that implements the `web.Controller` interface.

```go
type HomeController struct {
    tmpl *web.Templates
}

func (c *HomeController) Routes() []web.Route {
    return []web.Route{
        {Pattern: "GET /", Handler: c.Index},
        {Pattern: "GET /about", Handler: c.About},
    }
}

func (c *HomeController) Index(w http.ResponseWriter, r *http.Request) {
    c.tmpl.Render(w, http.StatusOK, "index.html", map[string]any{"Title": "Home Page"})
}

func (c *HomeController) About(w http.ResponseWriter, r *http.Request) {
    c.tmpl.Render(w, http.StatusOK, "about.html", nil)
}
```

### 2. Initialize and Run the Server

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"

    "github.com/meesooqa/go-web"
)

func main() {
    cfg := web.Config{
        Host:         "0.0.0.0",
        Port:         8080,
        TemplatesDir: "./templates",
        StaticDir:    "./static",
    }

    srv, err := web.New(cfg, slog.Default())
    if err != nil {
        panic(err)
    }

    // Register controllers
    homeCtrl := &HomeController{tmpl: srv.Templates()}
    srv.Register(homeCtrl)

    // Handle graceful shutdown
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    if err := srv.Run(ctx); err != nil {
        slog.Error("server error", "err", err)
    }
}
```

## Configuration

The server can be configured using the `web.Config` struct. You can pass this struct to `web.New()` or use `web.Load(path)` to load configuration from a YAML file.

| Field | Description | Default |
| :--- | :--- | :--- |
| `Host` | Address to listen on | `0.0.0.0` |
| `Port` | TCP port | `8080` |
| `ReadTimeout` | Max duration for reading request | `15s` |
| `ReadHeaderTimeout` | Max duration for reading headers | `5s` |
| `WriteTimeout` | Max duration for writing response | `15s` |
| `IdleTimeout` | Keep-alive connection timeout | `60s` |
| `ShutdownTimeout` | Graceful shutdown timeout | `10s` |
| `MaxHeaderBytes` | Max total size of request headers | `1MB` |
| `TemplatesDir` | Directory containing HTML templates | - |
| `StaticDir` | Directory for static files | - |
| `StaticURLPath` | URL prefix for static files | `/static/` |

## Examples

Check the `examples/` directory for more detailed usage scenarios.

## License

This project is licensed under the [LICENSE](LICENSE) file.
