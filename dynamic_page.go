package lago

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/components"
	"maragu.dev/gomponents"
)

// DynamicPage lazily resolves page elements by string identifiers from [RegistryPage] at build/render time.
// This decouples components registrations, avoiding import-time dependency loops between modular plugins.
//
// Use Cases:
//   - Lazy-loading sub-pages or sections dynamically from registries without establishing direct static dependencies.
//
// Example:
//
//	&lago.DynamicPage{
//	    Name: "admin.Dashboard",
//	}
type DynamicPage struct {
	// Page embeds common component properties like Key and Roles.
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

// GetChildren resolves the lazy target page from the registry and returns it in a slice.
func (d DynamicPage) GetChildren() []components.PageInterface {
	page, ok := RegistryPage.Get(d.Name)
	if !ok {
		slog.Warn("DynamicPage: page not found in registry", "name", d.Name)
		return nil
	}
	return []components.PageInterface{page}
}

// Build compiles the dynamic component by rendering the lazy resolved page.
func (d DynamicPage) Build(ctx context.Context) gomponents.Node {
	page, ok := RegistryPage.Get(d.Name)
	if !ok {
		slog.Warn("DynamicPage: page not found in registry", "name", d.Name)
		return nil
	}
	return components.Render(page, ctx)
}
