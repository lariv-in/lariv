package lariv

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/lariv-in/lariv/getters"
)

// Route represents a multiplexer-compatible HTTP routing entry.
//
// # Wildcard Path Guidelines
//
// 1. Wildcards (like `{id}`) must occupy a full path segment. Use `/users/u/{id}/` instead of `/users{id}/`.
// 2. Base paths should end with a trailing slash `/` before segment appends.
// 3. Sibling paths with fixed literals and wildcards under the same prefix should be disambiguated by adding an explicit sub-segment (e.g. `/users/roles/...` vs `/users/u/{id}/...`).
type Route struct {
	// Path represents the ServeMux-compatible URL path pattern (e.g., "/users/u/{id}/").
	Path string
	// Handler represents the HTTP handler mapped to this route path.
	Handler http.Handler
}

// RoutePath yields a Getter that resolves and interpolates path variables in a named route on the request's [App].
//
// Example:
//
//	pathGetter := RoutePath("user.profile", map[string]getters.Getter[any]{
//		"id": getters.Static(42),
//	})
//	url, err := pathGetter(ctx) // Resolves to "/users/u/42/profile"
func RoutePath(name string, args map[string]getters.Getter[any]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		app, ok := AppFromContext(ctx)
		if !ok || app == nil {
			return "", fmt.Errorf("Route for %s not found: App missing from context", name)
		}
		route, ok := app.Routes.Get(name)
		if !ok {
			return "", fmt.Errorf("Route for %s not found", name)
		}
		r := route.Path
		for k, g := range args {
			v, err := g(ctx)
			if err != nil {
				return "", err
			}
			r = strings.ReplaceAll(r, fmt.Sprintf("{%s}", k), fmt.Sprintf("%v", v))
		}
		return r, nil
	}
}
