// Package querypatchers contains explanations and code examples for database query patchers in Lariv.
//
// # Database Query Patchers (querypatchers.go)
//
// Query patchers decorate GORM query chains before database operations run (e.g. details, lists, deletes).
//
// # Existing Built-in Query Patchers
//
//   - views.QueryPatcherPreload[T]: eager-loads GORM associations to prevent N+1 queries.
//   - views.QueryPatcherSearch[T]: matches text search parameters to SQL LIKE/ILIKE structures.
//   - views.QueryPatcherJoinFilter[T]: joins associated database tables for relational filtering.
//   - views.QueryPatcherOrderBy[T]: orders GORM queries.
//
// # Creating a Custom Query Patcher
//
// Implement the views.QueryPatcher interface to apply custom SQL filters or scopes:
//
//	package myplugin
//
//	import (
//		"net/http"
//		"github.com/lariv-in/lariv/views"
//		"gorm.io/gorm"
//	)
//
//	type AccountFilterPatcher[T any] struct {
//		AccountID uint
//	}
//
//	func (p AccountFilterPatcher[T]) Patch(v views.View, r *http.Request, query gorm.ChainInterface[T]) gorm.ChainInterface[T] {
//		// Append custom multi-tenant SQL scope.
//		return query.Where("account_id = ?", p.AccountID)
//	}
//
//	// Registering inside LayerList or LayerDetail in views.go:
//	views.LayerList[BlogPost]{
//		QueryPatchers: views.QueryPatchers{
//			{Key: "tenant_filter", Value: AccountFilterPatcher[BlogPost]{AccountID: 42}},
//		},
//	}
package querypatchers
