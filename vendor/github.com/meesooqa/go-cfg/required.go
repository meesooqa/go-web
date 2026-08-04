package cfg

import (
	"reflect"
	"strings"
)

// findMissingRequired recursively searches for fields with the
// `required:"true"` tag that remained zero-valued after applying
// yaml/default/env values.
// prefix is used to build a readable path such as "database.dsn"
// in the error message.
func findMissingRequired(v reflect.Value, prefix string) []string {
	t := v.Type()
	var missing []string

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		name := fieldDisplayName(structField, prefix)

		if field.Kind() == reflect.Struct {
			missing = append(missing, findMissingRequired(field, name)...)
			continue
		}

		requiredTag, ok := structField.Tag.Lookup("required")
		if !ok || requiredTag != "true" {
			continue
		}

		if field.IsZero() {
			missing = append(missing, name)
		}
	}

	return missing
}

// fieldDisplayName builds a human-readable field name for an error message,
// using the name from the yaml tag if present, otherwise the struct field name.
func fieldDisplayName(structField reflect.StructField, prefix string) string {
	name := structField.Name
	if yamlTag, ok := structField.Tag.Lookup("yaml"); ok {
		yamlName := strings.Split(yamlTag, ",")[0]
		if yamlName != "" && yamlName != "-" {
			name = yamlName
		}
	}

	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
