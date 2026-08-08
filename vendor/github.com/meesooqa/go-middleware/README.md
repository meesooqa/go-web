# go-middleware

A lightweight, standard-compliant HTTP middleware library for Go.

`go-middleware` provides a set of essential middleware components designed to work seamlessly with the standard `net/http` package. It avoids custom abstractions, ensuring compatibility with any third-party middleware that follows the standard Go signature.

## Features

- **Middleware Chaining**: A simple utility to wrap handlers in a sequence of middleware layers.
- **Request Logging**: 
  - Logs request method, path, response status, duration, and remote address.
  - Automatic `X-Request-Id` generation and propagation for request tracing.
  - Integration with `github.com/meesooqa/go-lgr` to provide request-scoped loggers.
- **CORS Support**:
  - Configurable allowed origins, methods, and headers.
  - Intelligent handling of `AllowCredentials` and wildcard origins to comply with browser security specifications.
  - Built-in support for preflight `OPTIONS` requests.
- **Basic Authentication**:
  - Simple protection for administrative routes using `ADMIN_USER` and `ADMIN_PASS` environment variables.

## Installation

```bash
go get github.com/meesooqa/go-middleware
```

## Usage

### Middleware Signature and Chaining

The library uses the standard Go middleware pattern:
`type Middleware func(http.Handler) http.Handler`

You can use the `Chain` function to apply multiple middlewares to a handler. The first middleware in the list is the outermost layer.

```go
package main

import (
	"net/http"
	"github.com/meesooqa/go-middleware"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Request flow: Logging -> CORS -> handler
	// Response flow: handler -> CORS -> Logging
	chainedHandler := middleware.Chain(handler, 
		middleware.Logging(nil), 
		middleware.CORS(middleware.CORSConfig{
			AllowedOrigins: []string{"*"},
		}),
	)

	http.ListenAndServe(":8080", chainedHandler)
}
```

### Logging Middleware

The `Logging` middleware captures the lifecycle of an HTTP request. It integrates with `go-lgr` to attach a logger containing the `request_id` to the request context.

```go
// Use slog.Default() if no logger is provided
mw := middleware.Logging(nil) 
```

### CORS Middleware

The `CORS` middleware handles Cross-Origin Resource Sharing. It is configured via the `CORSConfig` struct.

```go
corsMw := middleware.CORS(middleware.CORSConfig{
    AllowedOrigins:   []string{"https://example.com", "https://api.example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
})
```

### Basic Authentication

The `BasicAuth` middleware protects routes using HTTP Basic Authentication. It is configured via the `BasicAuthConfig` struct, making it compatible with `github.com/meesooqa/go-cfg`.

Example of protecting only the `/admin` section:

```go
mux := http.NewServeMux()

// Public route
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Public Home"))
})

// Protected admin route
adminHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Welcome to the Admin Panel"))
})

// Define authentication config
authCfg := middleware.BasicAuthConfig{
    User:     "admin",
    Password: "secure-password",
}

// Wrap only the admin handler with BasicAuth
mux.Handle("/admin", middleware.BasicAuth(authCfg)(adminHandler))

http.ListenAndServe(":8080", mux)
```

## Configuration

The `CORSConfig` struct is compatible with [github.com/meesooqa/go-cfg](https://github.com/meesooqa/go-cfg), allowing you to load CORS settings directly from a YAML configuration file.

## License

This project is licensed under the [LICENSE](LICENSE) file.
