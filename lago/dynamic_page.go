package lago

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/components"
	"maragu.dev/gomponents"
)

// DynamicPage lazily resolves a page by name from RegistryPage at build time.
// This allows pages to reference other registered pages without import-time dependencies.
type DynamicPage struct {
	components.Page
	Name string
}

func (d DynamicPage) GetKey() string {
	return d.Key
}

func (d DynamicPage) GetRoles() []string {
	return d.Roles
}

func (d DynamicPage) GetChildren() []components.PageInterface {
	page, ok := RegistryPage.Get(d.Name)
	if !ok {
		slog.Warn("DynamicPage: page not found in registry", "name", d.Name)
		return nil
	}
	return []components.PageInterface{page}
}

func (d DynamicPage) Build(ctx context.Context) gomponents.Node {
	page, ok := RegistryPage.Get(d.Name)
	if !ok {
		slog.Warn("DynamicPage: page not found in registry", "name", d.Name)
		return nil
	}
	return components.Render(page, ctx)
}
