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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		record, ok := ctx.Value(key).(T)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		id := uint(reflect.ValueOf(record).FieldByName("ID").Uint())
		db := r.Context().Value("$db").(*gorm.DB)
		query := gorm.G[T](db).Where("id = ?", id)
		query = m.QueryPatchers.Apply(view, r, query)
		rowsAffected, err := query.Delete(ctx)
		fmt.Printf("Rows Affected: %d", rowsAffected)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		successUrl, err := m.SuccessURL(ctx)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, successUrl, http.StatusSeeOther)
	})
}
