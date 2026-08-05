package middleware

import "time"

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
