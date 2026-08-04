package cfg

import (
	"reflect"
	"strconv"
	"strings"
)

// setFieldValue converts the string raw into the type of field and sets
// its value. Supported types include string, signed integers, unsigned
// integers, float, bool, and []string (comma-separated values).
// Unsupported types and conversion errors are silently ignored — the field
// remains unchanged.
func setFieldValue(field reflect.Value, raw string) {
	if !field.CanSet() {
		return
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
			field.SetInt(n)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n, err := strconv.ParseUint(raw, 10, 64); err == nil {
			field.SetUint(n)
		}

	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(raw, 64); err == nil {
			field.SetFloat(f)
		}

	case reflect.Bool:
		if b, err := strconv.ParseBool(raw); err == nil {
			field.SetBool(b)
		}

	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(raw, ",")
			for i, p := range parts {
				parts[i] = strings.TrimSpace(p)
			}
			field.Set(reflect.ValueOf(parts))
		}
	}
}
