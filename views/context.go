package views

import (
	"context"
	"fmt"
	"maps"

	"github.com/lariv-in/lago/getters"
)

// ContextWithMap updates a map inside a context under the specified key.
// It retrieves the existing map from the context (or initializes a new one if missing),
// copies the provided keys and values into it using [maps.Copy], and returns the updated context.
func ContextWithMap[K comparable, V any](ctx context.Context, m map[K]V, key any) context.Context {
	ctxM, _ := ctx.Value(key).(map[K]V)
	if ctxM == nil {
		ctxM = map[K]V{}
	}
	maps.Copy(ctxM, m)
	return context.WithValue(ctx, key, ctxM)
}

// ContextWithErrorsAndValues merges input form parameter values and validation errors maps into the request context.
// It maps values to [getters.ContextKeyIn] and errors to [getters.ContextKeyError] so form input elements can retrieve them during render phases.
func ContextWithErrorsAndValues(ctx context.Context, values map[string]any, errors map[string]error) context.Context {
	return ContextWithMap(ContextWithMap(ctx, values, getters.ContextKeyIn), errors, getters.ContextKeyError)
}

type ContextNilPointerValueError[K any] struct {
	Key K
}

func (e ContextNilPointerValueError[K]) Error() string {
	return fmt.Sprintf("Value for key %#v was a nil pointer", e.Key)
}

type ContextTypeMismatchError[V any] struct {
	Value any
}

func (e ContextTypeMismatchError[V]) Error() string {
	var zero V
	return fmt.Sprintf("Type mismatch when getting value from context, expected %T, found %T with value %#v", zero, e.Value, e.Value)
}

func GetValueFromContext[K any, V any](ctx context.Context, key K) (V, error) {
	var zero V
	switch v := ctx.Value(key).(type) {
	case V:
		return v, nil
	case *V:
		if v == nil {
			return zero, ContextNilPointerValueError[K]{Key: key}
		}
		return *v, nil
	default:
		return zero, ContextTypeMismatchError[V]{
			Value: v,
		}
	}
}
