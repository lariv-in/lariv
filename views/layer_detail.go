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

// LayerDetail loads a single database record of type T by querying its primary key from a URL route path parameter.
// It acts as the primary data loader, storing the fetched record in the request context under Key for downstream layers or handlers (e.g. LayerUpdate or LayerDelete).
//
// QueryPatchers are applied to the query before execution, permitting preloading of associations, access control filtering, or tenant scopes.
//
// Use Cases:
//   - Fetching detail records for display views (e.g. profile edit pages, product specifications).
//   - Injecting model contexts for subsequent operations like record updates or deletions.
//
// Example:
//
//	views.View{
//	    Layers: []views.Layer{
//	        &views.PathLayer{Names: []string{"userId"}},
//	        views.LayerDetail[User]{
//	            PathParamKey: getters.Static("userId"),
//	            Key:          getters.Static("$user"),
//	            QueryPatchers: views.QueryPatchers{
//	                views.QueryPatcherPreload[User]("Profile"),
//	            },
//	        },
//	    },
//	}
type LayerDetail[T any] struct {
	// Key represents the context key string under which the loaded record instance is stored.
	// PathParamKey represents the URL path parameter name carrying the primary key (e.g., "id").
	Key, PathParamKey getters.Getter[string]

	// QueryPatchers represents the slice of query modifiers applied to GORM before retrieving the row.
	QueryPatchers QueryPatchers[T]
}

// Next wraps the downstream HTTP request handlers executing record loading.
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
