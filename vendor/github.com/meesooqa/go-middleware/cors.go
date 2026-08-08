package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CORSConfig defines the configuration for the CORS middleware.
// By default, AllowedOrigins is empty, meaning no CORS headers are added
// (a safe default: the browser applies the standard same-origin policy).
// To enable cross-origin requests, explicitly set AllowedOrigins
//
// Struct tags are compatible with `github.com/meesooqa/go-cfg`
type CORSConfig struct {
	// AllowedOrigins is a list of permitted origins, or ["*"] to allow all.
	// Empty by default (CORS is disabled)
	AllowedOrigins []string `yaml:"allowed_origins"`

	// AllowedMethods is a list of HTTP methods permitted in the preflight response
	AllowedMethods []string `yaml:"allowed_methods" default:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`

	// AllowedHeaders is a list of headers permitted in the preflight response
	AllowedHeaders []string `yaml:"allowed_headers" default:"Content-Type,Authorization"`

	// AllowCredentials allows sending credentials (cookies, Authorization)
	// in cross-origin requests. See the warning in CORS() —
	// this is incompatible with AllowedOrigins: ["*"] per the specification
	AllowCredentials bool `yaml:"allow_credentials" default:"false"`

	// MaxAge specifies how long the browser can cache the preflight request response
	MaxAge time.Duration `yaml:"max_age" default:"12h"`
}

// CORS creates a middleware that adds CORS headers based on c.
// If c.AllowedOrigins is empty, the middleware is a no-op;
// it simply passes the request through without adding any headers.
//
// An important nuance of the CORS specification:
// the combination of AllowedOrigins: ["*"] and AllowCredentials: true
// is prohibited by browsers; a server must not return
// Access-Control-Allow-Origin: * alongside Access-Control-Allow-Credentials: true.
// In this case, the middleware automatically falls back to "echoing"
// the specific request origin instead of using "*", ensuring the configuration
// remains functional rather than being silently rejected by the browser.
func CORS(c CORSConfig) Middleware {
	wildcard := containsString(c.AllowedOrigins, "*")
	allowedMethods := strings.Join(c.AllowedMethods, ", ")
	allowedHeaders := strings.Join(c.AllowedHeaders, ", ")
	maxAge := strconv.Itoa(int(c.MaxAge.Seconds()))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" && (wildcard || containsString(c.AllowedOrigins, origin)) {
				if wildcard && !c.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					// Returning a specific origin instead of "*" is mandatory
					// when AllowCredentials is true, and simply safer otherwise
					// (it avoids granting broader access than requested).
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Add("Vary", "Origin")
				}

				if c.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				if allowedMethods != "" {
					w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
				}
				if allowedHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
				}
				if c.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", maxAge)
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func containsString(list []string, target string) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}
