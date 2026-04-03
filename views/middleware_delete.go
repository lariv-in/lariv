package views

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

type MiddlewareDelete[T any] struct {
	Key           getters.Getter[string]
	SuccessURL    getters.Getter[string]
	PatchParamKey getters.Getter[string]
	QueryPatchers QueryPatchers[T]
}

func (m MiddlewareDelete[T]) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		key, err := m.Key(ctx)
		if err != nil {
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": fmt.Errorf("failed to resolve context key: %w", err),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		record, ok := ctx.Value(key).(T)
		if !ok {
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": fmt.Errorf("record not found in context"),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		id := uint(reflect.ValueOf(record).FieldByName("ID").Uint())
		db := ctx.Value("$db").(*gorm.DB)
		query := gorm.G[T](db).Where("id = ?", id)
		query = m.QueryPatchers.Apply(view, r, query)
		_, err = query.Delete(ctx)
		if err != nil {
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
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": fmt.Errorf("failed to resolve redirect URL: %w", err),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		http.Redirect(w, r, successUrl, http.StatusSeeOther)
	})
}
