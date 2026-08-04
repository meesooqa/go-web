package cfg

import "reflect"

// applyDefaults recursively walks through the fields of struct v and applies
// values from the `default:"..."` tag if a field is still zero-valued
// (was not specified in the YAML).
func applyDefaults(v reflect.Value) {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		if field.Kind() == reflect.Struct {
			applyDefaults(field)
			continue
		}

		defaultTag, ok := structField.Tag.Lookup("default")
		if !ok {
			continue
		}

		if !field.IsZero() {
			continue
		}

		setFieldValue(field, defaultTag)
	}
}
