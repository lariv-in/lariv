package views

import (
	"net/http"

	"gorm.io/gorm"
)

// QueryPatcherOrderBy applies a default SQL ORDER BY sort clause to a GORM query chain.
//
// Use Cases:
//   - Enforcing default sorting orders on index tables (e.g. listing elements by creation dates descending, or alphabetically).
//
// Example:
//
//	views.View{
//	    Layers: []views.Layer{
//	        views.LayerList[Product]{
//	            QueryPatchers: views.QueryPatchers{
//	                views.NewPair("default_sort", views.QueryPatcherOrderBy[Product]{
//	                    Order: "created_at DESC",
//	                }),
//	            },
//	        },
//	    },
//	}
type QueryPatcherOrderBy[T any] struct {
	// Order represents the SQL order clause string (e.g., "name ASC", "id DESC").
	Order string
}

// Patch applies the ORDER BY statement to the GORM query chain.
func (p QueryPatcherOrderBy[T]) Patch(_ View, _ *http.Request, db gorm.ChainInterface[T]) gorm.ChainInterface[T] {
	return db.Order(p.Order)
}
