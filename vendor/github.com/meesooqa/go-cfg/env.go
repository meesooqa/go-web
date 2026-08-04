package cfg

import (
	"fmt"
	"os"
	"reflect"

	"github.com/joho/godotenv"
)

// applyEnv recursively walks through the fields of struct v and overrides
// values using the environment variable specified in the `env:"VAR_NAME"` tag,
// if such a variable is set and not empty.
func applyEnv(v reflect.Value) {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		if field.Kind() == reflect.Struct {
			applyEnv(field)
			continue
		}

		envTag, ok := structField.Tag.Lookup("env")
		if !ok {
			continue
		}

		envValue, exists := os.LookupEnv(envTag)
		if !exists || envValue == "" {
			continue
		}

		setFieldValue(field, envValue)
	}
}

// loadEnvFiles loads variables from .env files into the process
// environment. A missing file is skipped without an error. Actual
// environment variables already set in the process are not overwritten.
// If multiple files are provided, the first file in the list takes precedence.
func loadEnvFiles(paths ...string) error {
	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			continue
		}

		if err := godotenv.Load(p); err != nil {
			return fmt.Errorf("cfg: failed to load .env file %q: %w", p, err)
		}
	}
	return nil
}
