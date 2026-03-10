package components

import "reflect"

// MapFromStruct converts a struct into a map[string]any using reflection.
func MapFromStruct(s any) map[string]any {
	if m, ok := s.(map[string]any); ok {
		return m
	}

	m := make(map[string]any)
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return m
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // Skip unexported fields
			continue
		}
		m[field.Name] = v.Field(i).Interface()
	}
	return m
}
