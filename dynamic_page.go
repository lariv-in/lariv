package lariv

import (
	"context"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv/components"
)

// DynamicPage lazily resolves page elements by string identifiers from the compiled [App] catalog at build/render time.
// This decouples component registrations, avoiding import-time dependency loops between modular plugins.
type DynamicPage struct {
	components.Page
	// Name represents the registered string identifier of the target page component to fetch (e.g. "admin.Dashboard").
	Name string
}

// GetKey returns the unique key identifier for this DynamicPage component.
func (d DynamicPage) GetKey() string {
	return d.Key
}

// GetRoles returns the authorized roles required to view this DynamicPage.
func (d DynamicPage) GetRoles() []string {
	return d.Roles
}

// GetChildren resolves the lazy target page from the catalog and returns it in a slice.
// Without a catalog argument this returns nil; prefer Build for rendering.
func (d DynamicPage) GetChildren() []components.PageInterface {
	return nil
}

// Build compiles the dynamic component by rendering the lazy resolved page from cat.
func (d DynamicPage) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	page, ok := cat.Page(d.Name)
	if !ok {
		slog.Warn("DynamicPage: page not found in catalog", "name", d.Name)
		return nil
	}
	return components.Render(page, cat, ctx, w)
}
