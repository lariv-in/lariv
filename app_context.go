package lariv

import (
	"context"
	"fmt"

	"github.com/lariv-in/lariv/components"
)

type appContextKey struct{}

// ContextWithApp returns a child context carrying the compiled [App] and its [components.Catalog].
func ContextWithApp(ctx context.Context, app *App) context.Context {
	ctx = context.WithValue(ctx, appContextKey{}, app)
	return components.ContextWithCatalog(ctx, app)
}

// AppFromContext returns the [App] stored by [ContextWithApp], if any.
func AppFromContext(ctx context.Context) (*App, bool) {
	app, ok := ctx.Value(appContextKey{}).(*App)
	return app, ok
}

// MustAppFromContext returns the [App] from ctx or panics.
func MustAppFromContext(ctx context.Context) *App {
	app, ok := AppFromContext(ctx)
	if !ok || app == nil {
		panic(fmt.Errorf("lariv: App missing from context"))
	}
	return app
}
