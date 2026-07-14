package views

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

// LayerDelete handles database row removal operations for type T on DELETE/POST request actions.
// It expects the target record to already reside inside the context under Key (typically placed by a preceding [LayerDetail] layer).
//
// Upon intercepting a delete trigger, it extracts the record's primary ID, runs GORM deletions through any configured [QueryPatchers],
// and handles redirections or downstream handler execution depending on SuccessURL mappings.
//
// Use Cases:
//   - Supporting resource deletion handlers (e.g. deleting users, clearing obsolete transactions).
//   - Applying scope-based delete queries (e.g., confirming record ownership before issuing DB delete commands).
//
// Example:
//
//	views.View{
//	    Layers: []views.Layer{
//	        views.LayerDetail[User]{Key: getters.Static("$record")},
//	        views.LayerDelete[User]{
//	            Key:        getters.Static("$record"),
//	            SuccessURL: lago.RoutePath("users.List", nil),
//	        },
//	    },
//	}
type LayerDelete[T any] struct {
	// Key is the Getter function returning the context key pointing to the target record to delete.
	Key getters.Getter[string]
	// SuccessURL represents the dynamic Getter resolving to the redirection target URL upon successful deletion.
	SuccessURL getters.Getter[string]
	// QueryPatchers represents the slice of query modifications to restrict deletion scopes.
	QueryPatchers QueryPatchers[T]
}

// Next wraps the downstream HTTP request handlers executing row deletions.
func (m LayerDelete[T]) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("views: layer delete: resolve context key", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": fmt.Errorf("failed to resolve context key: %w", err),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		record, ok := ctx.Value(key).(T)
		if !ok {
			slog.Error("views: layer delete: record missing from context", "key", key)
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": fmt.Errorf("record not found in context"),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		id := uint(reflect.ValueOf(record).FieldByName("ID").Uint())
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("views: layer delete: db from context", "error", dberr)
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": dberr,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		query := gorm.G[T](db).Where("id = ?", id)
		query = m.QueryPatchers.Apply(view, r, query)
		_, err = query.Delete(ctx)
		if err != nil {
			slog.Error("views: layer delete: delete record", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": fmt.Errorf("failed to delete record: %w", err),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if m.SuccessURL == nil {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successUrl, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("views: layer delete: resolve success URL", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": fmt.Errorf("failed to resolve redirect URL: %w", err),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		HtmxRedirect(w, r, successUrl, http.StatusSeeOther)
	})
}
