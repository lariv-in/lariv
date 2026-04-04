package views

import (
	"net/http"

	"gorm.io/gorm"
)

// QueryPatcherOrderBy applies ORDER BY (e.g. "name ASC").
type QueryPatcherOrderBy[T any] struct {
	Order string
}

func (p QueryPatcherOrderBy[T]) Patch(_ View, _ *http.Request, db gorm.ChainInterface[T]) gorm.ChainInterface[T] {
	return db.Order(p.Order)
}
