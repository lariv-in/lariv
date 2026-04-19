package views

import (
	"net/http"

	"gorm.io/gorm"
)

// QueryPatcherPreload preloads associations on the list/detail query (GORM association names / dotted paths).
type QueryPatcherPreload[T any] struct {
	Fields         []string
	PreloadBuilder func(View, *http.Request, gorm.PreloadBuilder) error
}

func (p QueryPatcherPreload[T]) Patch(v View, r *http.Request, db gorm.ChainInterface[T]) gorm.ChainInterface[T] {
	var preloadBuilder func(gorm.PreloadBuilder) error

	if p.PreloadBuilder != nil {
		preloadBuilder = func(pb gorm.PreloadBuilder) error {
			return p.PreloadBuilder(v, r, pb)
		}
	}

	for _, f := range p.Fields {
		db = db.Preload(f, preloadBuilder)
	}
	return db
}
