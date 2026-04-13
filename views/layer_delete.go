package views

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

// LayerDelete handles row deletion for type T on DELETE requests.
// On non-DELETE methods it passes through to next unchanged.
//
// It expects the record to already be in the context under Key, typically
// placed there by a preceding LayerDetail. On DELETE it extracts the
// record's ID and issues a DELETE query with any QueryPatchers applied.
//
// If SuccessURL is set, a successful delete redirects to the resolved URL.
// If SuccessURL is nil, next is called so a downstream handler can decide
// the response.
//
// All errors (missing context record, DB failures, getter failures) are placed
// into getters.ContextKeyError under "_global" and next is called, never a raw
// HTTP response.
type LayerDelete[T any] struct {
	Key           getters.Getter[string]
	SuccessURL    getters.Getter[string]
	QueryPatchers QueryPatchers[T]
}

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
