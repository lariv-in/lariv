package getters

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// contextMapStep returns m[part] or nil if the key is missing (for Key path traversal).
func contextMapStep[V any](m map[string]V, part string) any {
	if v, ok := m[part]; ok {
		return v
	}
	return nil
}

// Key returns a Getter that gets the value from the context.
// '.' can be used to traverse map or struct fields. Keys must match exactly.
// Returns the zero value of T when key is not found, with an error
func Key[T any](key string) Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		parts := strings.Split(key, ".")
		value := ctx.Value(parts[0])
		for _, part := range parts[1:] {
			if value == nil {
				return zero, fmt.Errorf("Couldn't find %s in context", key)
			}
			switch m := value.(type) {
			case map[string]any:
				value = contextMapStep(m, part)
			case map[string]error:
				value = contextMapStep(m, part)
			default:
				v, ok := value.(reflect.Value)
				if !ok {
					v = reflect.ValueOf(value)
				}
				if v.Kind() == reflect.Pointer {
					v = v.Elem()
				}
				flat := MapFromStruct(v)
				if v, exists := flat[part]; exists {
					value = v
				} else {
					value = nil
				}
			}
		}
		if value == nil {
			return zero, nil
		}
		v, ok := value.(T)
		if !ok {
			return zero, fmt.Errorf("Value for key %s found, but the type of value in context was %v, expected %v", key, reflect.TypeOf(value), reflect.TypeOf(zero))
		}
		return v, nil
	}
}
