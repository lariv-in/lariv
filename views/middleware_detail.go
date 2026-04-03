package views

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

type MiddlewareDetail[T any] struct {
	Key, PathParamKey getters.Getter[string]

	QueryPatchers QueryPatchers[T]
}

func (m MiddlewareDetail[T]) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			pathParamKey, err := m.PathParamKey(ctx)
			if err != nil {
				ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
					"_global": fmt.Errorf("failed to resolve path param key: %w", err),
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			idStr := r.PathValue(pathParamKey)
			id, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
					"_global": fmt.Errorf("invalid ID %q", idStr),
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			db := ctx.Value("$db").(*gorm.DB)
			query := m.QueryPatchers.Apply(view, r, gorm.G[T](db).Scopes())
			instance, err := query.Where("ID = ?", id).First(ctx)
			if err != nil {
				ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
					"_global": fmt.Errorf("record not found"),
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			key, err := m.Key(ctx)
			if err != nil {
				ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
					"_global": fmt.Errorf("failed to resolve context key: %w", err),
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			ctx = context.WithValue(ctx, key, instance)
			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}
