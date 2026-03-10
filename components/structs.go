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

	flattenStruct(v, m)
	return m
}

func flattenStruct(v reflect.Value, m map[string]any) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		if field.Anonymous && v.Field(i).Kind() == reflect.Struct {
			flattenStruct(v.Field(i), m)
			continue
		}
		m[field.Name] = v.Field(i).Interface()
	}
}
