package views

import (
	"net/http"

	"gorm.io/gorm"
)

// QueryPatcherPreload preloads associations on the list/detail query (GORM association names / dotted paths).
type QueryPatcherPreload[T any] struct {
	Fields []string
}

func (p QueryPatcherPreload[T]) Patch(_ View, _ *http.Request, db gorm.ChainInterface[T]) gorm.ChainInterface[T] {
	for _, f := range p.Fields {
		db = db.Preload(f, nil)
	}
	return db
}
