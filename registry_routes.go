package lariv

import (
	"context"
	"fmt"
	"net/http"
	pprof_http "net/http/pprof"
	"runtime/pprof"
	"strings"

	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

// RegistryRoute represents the global immutable registry tracking route mappings.
var RegistryRoute *registry.ImmutableRegistry[Route] = &registry.ImmutableRegistry[Route]{}

// Route represents a multiplexer-compatible HTTP routing entry.
//
// # Wildcard Path Guidelines
//
// 1. Wildcards (like `{id}`) must occupy a full path segment. Use `/users/u/{id}/` instead of `/users{id}/`.
// 2. Base paths should end with a trailing slash `/` before segment appends.
// 3. Sibling paths with fixed literals and wildcards under the same prefix should be disambiguated by adding an explicit sub-segment (e.g. `/users/roles/...` vs `/users/u/{id}/...`).
//
// Use Cases:
//   - Mapping standard HTTP endpoints to custom controllers or view handlers.
//
// Example Definition:
//
//	var ProfileRoute = Route{
//		Path: "/users/u/{id}/profile",
//		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			id := r.PathValue("id")
//			w.Write([]byte("Profile for user: " + id))
//		}),
//	}
//
// Example Registration:
//
//	// In your lariv.Plugin setup:
//	lariv.Plugin{
//		Routes: lariv.PluginStages(func() PluginFeatures[Route] {
//			return PluginFeatures[Route]{
//				Entries: []registry.Pair[string, Route]{
//					registry.NewPair("user.profile", ProfileRoute),
//				},
//			}
//		}),
//	}
//
// Example Patch:
//
//	// Register a patch to prefix or redirect routes from another plugin:
//	lariv.Plugin{
//		Routes: lariv.PluginStages(func() PluginFeatures[Route] {
//			return PluginFeatures[Route]{
//				Patches: []registry.Pair[string, func(Route) Route]{
//					registry.NewPair("user.profile", func(existing Route) Route {
//						return Route{
//							Path:    "/internal" + existing.Path,
//							Handler: existing.Handler,
//						}
//					}),
//				},
//			}
//		}),
//	}
//
// Example Retrieval:
//
//	route, ok := RegistryRoute.Get("user.profile")
type Route struct {
	// Path represents the ServeMux-compatible URL path pattern (e.g., "/users/u/{id}/").
	Path string
	// Handler represents the HTTP handler mapped to this route path.
	Handler http.Handler
}

// GetRouter initializes and returns the ServeMux router populated with all registered routes.
// It maps path endings to specific ServeMux rules and registers pprof handlers under debug environments.
func GetRouter(config LarivConfig) *http.ServeMux {
	baseRouter := http.NewServeMux()
	if config.Debug {
		baseRouter.HandleFunc("/pprof/", pprof_http.Index)
		fmt.Printf("Added debug route for profile index at /pprof/\n")
		baseRouter.HandleFunc("/pprof/cmdline/", pprof_http.Cmdline)
		fmt.Printf("Added debug route for profile cmdline index at /pprof/cmdline/\n")
		baseRouter.HandleFunc("/pprof/profile/", pprof_http.Profile)
		fmt.Printf("Added debug route for profile 'profile' index at /pprof/profile/\n")
		baseRouter.HandleFunc("/pprof/symbol/", pprof_http.Symbol)
		fmt.Printf("Added debug route for profile symbol index at /pprof/symbol/\n")
		baseRouter.HandleFunc("/pprof/trace/", pprof_http.Trace)
		fmt.Printf("Added debug route for profile trace index at /pprof/trace/\n")
		for _, profile := range pprof.Profiles() {
			profile_name := profile.Name()
			profile_route := fmt.Sprintf("/pprof/%s/", profile_name)
			baseRouter.Handle(profile_route, pprof_http.Handler(profile.Name()))
			fmt.Printf("Added debug route for profile %s at %s\n", profile_name, profile_route)
		}
	}
	routes := RegistryRoute.All()
	for _, route := range routes {
		// Keep exact-match behavior for "directory-like" routes that end with "/"
		// (so "/foo/" doesn't also match "/foo/bar"). For non-slash routes like
		// "/app.webmanifest", register them directly since "/app.webmanifest{$}"
		// is not a valid ServeMux pattern.
		if strings.HasSuffix(route.Path, "/") {
			baseRouter.Handle(route.Path+"{$}", route.Handler)
		} else {
			baseRouter.Handle(route.Path, route.Handler)
		}
	}
	return baseRouter
}

// RoutePath yields a Getter that resolves and interpolates path variables in a registered named route.
//
// Example:
//
//	pathGetter := RoutePath("user.profile", map[string]getters.Getter[any]{
//		"id": getters.Static(42),
//	})
//	url, err := pathGetter(ctx) // Resolves to "/users/u/42/profile"
func RoutePath(name string, args map[string]getters.Getter[any]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		if route, ok := RegistryRoute.Get(name); ok {
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
		return "", fmt.Errorf("Route for %s not found", name)
	}
}
