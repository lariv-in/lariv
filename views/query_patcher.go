package views

import (
	"net/http"

	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

type QueryPatchers[T any] []registry.Pair[string, QueryPatcher[T]]

func (q QueryPatchers[T]) Apply(view View, r *http.Request, query gorm.ChainInterface[T]) gorm.ChainInterface[T] {
	for _, queryPatcher := range q {
		query = queryPatcher.Value.Patch(view, r, query)
	}
	return query
}

type QueryPatcher[T any] interface {
	Patch(View, *http.Request, gorm.ChainInterface[T]) gorm.ChainInterface[T]
}
