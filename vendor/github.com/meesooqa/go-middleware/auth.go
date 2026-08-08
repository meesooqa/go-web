package middleware

import (
	"net/http"
)

// BasicAuthConfig defines the configuration for the BasicAuth middleware.
//
// Struct tags are compatible with `github.com/meesooqa/go-cfg`
type BasicAuthConfig struct {
	// User is the username required for authentication.
	User string `env:"ADMIN_USER"`

	// Password is the password required for authentication.
	Password string `env:"ADMIN_PASS"`
}

// BasicAuth is a middleware that protects routes using HTTP Basic Authentication.
// It verifies the request credentials against the provided configuration.
// If authentication fails, it returns a 401 Unauthorized response with the WWW-Authenticate header,
// prompting the browser for credentials.
func BasicAuth(cfg BasicAuthConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()

			if !ok || user != cfg.User || pass != cfg.Password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
