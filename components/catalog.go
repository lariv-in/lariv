package components

import "context"

// NamedSidebarItem pairs a registry key with a [SidebarItem] for right-drawer tabs.
type NamedSidebarItem struct {
	Key  string
	Item SidebarItem
}

// Catalog is the narrow, read-only surface components need from a compiled app.
// lariv.App implements this interface; components must not import lariv.
type Catalog interface {
	Page(name string) (PageInterface, bool)
	HeadNodes() []PageInterface
	TopbarItems() []PageInterface
	RightSidebarItems() []NamedSidebarItem
}

// EmptyCatalog is a no-op Catalog for tests and leaf renders that do not need lookups.
type EmptyCatalog struct{}

func (EmptyCatalog) Page(string) (PageInterface, bool)     { return nil, false }
func (EmptyCatalog) HeadNodes() []PageInterface            { return nil }
func (EmptyCatalog) TopbarItems() []PageInterface          { return nil }
func (EmptyCatalog) RightSidebarItems() []NamedSidebarItem { return nil }

type catalogContextKey struct{}

// ContextWithCatalog returns a child context carrying the compiled [Catalog].
func ContextWithCatalog(ctx context.Context, cat Catalog) context.Context {
	return context.WithValue(ctx, catalogContextKey{}, cat)
}

// CatalogFromContext returns the [Catalog] stored by [ContextWithCatalog], or [EmptyCatalog].
func CatalogFromContext(ctx context.Context) Catalog {
	if cat, ok := ctx.Value(catalogContextKey{}).(Catalog); ok && cat != nil {
		return cat
	}
	return EmptyCatalog{}
}
