package views

import (
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// QueryPatcherSearch applies a case-insensitive OR contains filter (using ILIKE) across multiple database columns.
// It executes when the "search" URL query parameter is non-empty (e.g. ?search=term).
//
// Use Cases:
//   - Implementing global search input controls on data index tables to search records by text fields (e.g., matching users by first name, last name, or email).
//
// Example:
//
//	views.View{
//	    Layers: []views.Layer{
//	        views.LayerList[User]{
//	            QueryPatchers: views.QueryPatchers{
//	                views.NewPair("global_search", views.QueryPatcherSearch[User]{
//	                    Columns: []string{"first_name", "last_name", "email"},
//	                }),
//	            },
//	        },
//	    },
//	}
type QueryPatcherSearch[T any] struct {
	// Columns represents the list of database column names to match using OR ILIKE clauses.
	Columns []string
}

// Patch applies the OR ILIKE search clauses to the GORM query chain.
func (p QueryPatcherSearch[T]) Patch(_ View, r *http.Request, db gorm.ChainInterface[T]) gorm.ChainInterface[T] {
	term := strings.TrimSpace(r.URL.Query().Get("search"))
	if term == "" || len(p.Columns) == 0 {
		return db
	}

	pattern := "%" + term + "%"
	clauses := make([]string, len(p.Columns))
	args := make([]any, len(p.Columns))
	for i, col := range p.Columns {
		clauses[i] = col + " ILIKE ?"
		args[i] = pattern
	}
	return db.Where(strings.Join(clauses, " OR "), args...)
}
