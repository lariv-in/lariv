package views

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

// LayerDetail fetches a single record of type T by its primary key from a URL
// path parameter and stores it in the request context under Key.
//
// This layer is the sole owner of "load one record by URL PK" logic. Other
// layers that need the record (LayerUpdate, LayerDelete) expect it
// to already be in the context and should be composed after LayerDetail.
//
// PathParamKey resolves to the name of the URL path parameter that carries the
// primary key (e.g. "id"). Key resolves to the context key under which the
// loaded T instance is stored for downstream handlers.
//
// QueryPatchers are applied to the query before executing it, allowing callers
// to add preloads, scopes, or tenant filters.
//
// On any error (bad path param, record not found, getter failure) the layer
// sets a "_global" error in getters.ContextKeyError and calls next instead of
// writing an HTTP response directly.
type LayerDetail[T any] struct {
	Key, PathParamKey getters.Getter[string]

	QueryPatchers QueryPatchers[T]
}

func (m LayerDetail[T]) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			pathParamKey, err := m.PathParamKey(ctx)
			if err != nil {
				slog.Error("views: layer detail: resolve path param key", "error", err)
				ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
					"_global": fmt.Errorf("failed to resolve path param key: %w", err),
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			idStr := r.PathValue(pathParamKey)
			id, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				slog.Error("views: layer detail: parse id", "error", err, "id", idStr)
				ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
					"_global": fmt.Errorf("invalid ID %q", idStr),
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			db, dberr := getters.DBFromContext(ctx)
			if dberr != nil {
				slog.Error("views: layer detail: db from context", "error", dberr)
				ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
					"_global": dberr,
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			query := m.QueryPatchers.Apply(view, r, gorm.G[T](db).Scopes())
			instance, err := query.Where("ID = ?", id).First(ctx)
			if err != nil {
				slog.Error("views: layer detail: load record", "error", err, "id", id)
				ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
					"_global": fmt.Errorf("record not found"),
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			key, err := m.Key(ctx)
			if err != nil {
				slog.Error("views: layer detail: resolve context key", "error", err)
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
