package views

import (
	"context"
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
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			idStr := r.PathValue(pathParamKey)
			id, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			db := r.Context().Value("$db").(*gorm.DB)
			query := m.QueryPatchers.Apply(view, r, gorm.G[T](db).Scopes())
			instance, err := query.Where("ID = ?", id).First(r.Context())
			if err != nil {
				http.NotFound(w, r)
				return
			}
			key, err := m.Key(ctx)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			ctx = context.WithValue(ctx, key, instance)
			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}
