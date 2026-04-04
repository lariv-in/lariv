package views

import (
	"net/http"

	"gorm.io/gorm"
)

// QueryPatcherPreload preloads an association on the list/detail query.
type QueryPatcherPreload[T any] struct {
	Field string
}

func (p QueryPatcherPreload[T]) Patch(_ View, _ *http.Request, db gorm.ChainInterface[T]) gorm.ChainInterface[T] {
	return db.Preload(p.Field, nil)
}
