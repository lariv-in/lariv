package lago

import (
	"context"
	"log/slog"

	"maragu.dev/gomponents"
)

// DynamicPage lazily resolves a page by name from RegistryPage at build time.
// This allows pages to reference other registered pages without import-time dependencies.
type DynamicPage struct {
	Name string
}

func (d DynamicPage) Build(ctx context.Context) gomponents.Node {
	page, ok := RegistryPage.Get(d.Name)
	if !ok {
		slog.Warn("DynamicPage: page not found in registry", "name", d.Name)
		return nil
	}
	return page.Build(ctx)
}
