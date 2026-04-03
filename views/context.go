package views

import (
	"context"
	"maps"

	"github.com/lariv-in/lago/getters"
)

func ContextWithMap[K comparable, V any](ctx context.Context, m map[K]V, key any) context.Context {
	ctxM, _ := ctx.Value(key).(map[K]V)
	if ctxM == nil {
		ctxM = map[K]V{}
	}
	maps.Copy(ctxM, m)
	return context.WithValue(ctx, key, ctxM)
}

func ContextWithErrorsAndValues(ctx context.Context, values map[string]any, errors map[string]error) context.Context {
	return ContextWithMap(ContextWithMap(ctx, values, getters.ContextKeyIn), errors, getters.ContextKeyError)
}
