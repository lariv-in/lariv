package views

import (
	"net/http"

	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// QueryPatchers represents a sequence of QueryPatcher registry pairs applied in order.
type QueryPatchers[T any] []registry.Pair[string, QueryPatcher[T]]

// Apply executes all registered query patcher rules, chaining and returning the resulting GORM interface.
func (q QueryPatchers[T]) Apply(view View, r *http.Request, query gorm.ChainInterface[T]) gorm.ChainInterface[T] {
	for _, queryPatcher := range q {
		query = queryPatcher.Value.Patch(view, r, query)
	}
	return query
}

// QueryPatcher defines an interface for component utilities capable of modifying a database query chain.
// It is applied during database retrieval flows (like detail queries, list queries, updates, or deletions) to append SQL filters.
//
// Use Cases:
//   - Injecting preloads to eager-load relations (e.g. preloading profiles or tags).
//   - Appending multi-tenant scoping filters (e.g., scoping data queries by current account ID).
//   - Enforcing state constraints (e.g. filtering out soft-deleted items or restricting lists to active states).
//
// Example:
//
//	type TenantScopePatcher struct{}
//
//	func (p TenantScopePatcher) Patch(view views.View, r *http.Request, query gorm.ChainInterface[Product]) gorm.ChainInterface[Product] {
//		tenantID := GetTenantIDFromContext(r.Context())
//		return query.Where("tenant_id = ?", tenantID)
//	}
type QueryPatcher[T any] interface {
	// Patch injects queries, preloads, scopes, or joins, returning the decorated GORM chain.
	Patch(View, *http.Request, gorm.ChainInterface[T]) gorm.ChainInterface[T]
}
