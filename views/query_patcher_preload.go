package views

import (
	"net/http"

	"gorm.io/gorm"
)

// QueryPatcherPreload applies GORM eager preloads to the database query chain, fetching related association entities.
// Eager preloads prevent N+1 database queries when rendering rows and grids referencing nested relational data fields.
//
// Use Cases:
//   - Eager-loading relational dependencies (e.g. preloading user profiles, loading item tags).
//   - Customizing relation preloads (e.g. ordering preloaded history logs chronologically).
//
// Example:
//
//	views.View{
//	    Layers: []views.Layer{
//	        views.LayerList[Product]{
//	            QueryPatchers: views.QueryPatchers{
//	                views.NewPair("preload_tags", views.QueryPatcherPreload[Product]{
//	                    Fields: []string{"Tags"},
//	                }),
//	            },
//	        },
//	    },
//	}
type QueryPatcherPreload[T any] struct {
	// Fields represents the slice of GORM association names or nested dotted paths to eager load (e.g. "Profile", "Profile.Address").
	Fields         []string
	// PreloadBuilder represents an optional callback allowing custom SQL builders on relation queries.
	PreloadBuilder func(View, *http.Request, gorm.PreloadBuilder) error
}

// Patch applies eager-load preloads for the registered fields to the GORM query chain.
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
