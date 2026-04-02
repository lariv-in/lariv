package getters

import (
	"reflect"
)

// MapFromStruct converts a struct into a map[string]any using reflection.
func MapFromStruct(s any) map[string]any {
	if m, ok := s.(map[string]any); ok {
		return m
	}

	m := make(map[string]any)
	var v reflect.Value
	v, ok := s.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(s)
	}

	if !v.IsValid() {
		return m
	}

	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return m
		}
		v = v.Elem()
	}

	if !v.IsValid() {
		return m
	}

	if m, ok := v.Interface().(map[string]any); ok {
		return m
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
		if v.Field(i).Kind() == reflect.Struct {
			for k, v := range MapFromStruct(v.Field(i)) {
				m[field.Name+"."+k] = v
			}
		}
		m[field.Name] = v.Field(i).Interface()
	}
}
