// Package cfg provides a generic configuration loader
// for Go applications. Settings are loaded from a YAML file,
// supplemented with default values (the `default` tag),
// overridden by environment variables (the `env` tag, optionally
// loaded from .env files), and validated for required fields
// (the `required:"true"` tag).
//
// The package does not define any application-specific structures—
// the configuration schema (type T) is defined by the application
// that uses cfg.
package cfg

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load reads the YAML file at the specified path and any optional
// .env files, parses them into a value of type T, applies default
// values and environment variables, and then validates required fields.
//
// Value source precedence: env > YAML > default.
//
// envFiles is an optional list of paths to .env files. If a file
// does not exist, it is silently skipped. Actual process environment
// variables always take precedence over values loaded from .env files.
func Load[T any](path string, envFiles ...string) (*T, error) {
	if err := loadEnvFiles(envFiles...); err != nil {
		return nil, err
	}

	var cfg T

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cfg: failed to read config %q: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cfg: failed to parse config %q: %w", path, err)
	}

	root := reflect.ValueOf(&cfg).Elem()

	applyDefaults(root)
	applyEnv(root)

	if missing := findMissingRequired(root, ""); len(missing) > 0 {
		return nil, fmt.Errorf("cfg: missing required configuration fields: %s", strings.Join(missing, ", "))
	}

	return &cfg, nil
}
