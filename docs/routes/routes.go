// Package routes contains explanations and code examples for HTTP routing in Lariv.
//
// # HTTP Routing (routes.go)
//
// Lariv routes map request path patterns directly to handlers.
// The routing layer is built on top of Go 1.22+'s native net/http router (ServeMux), supporting path variables natively.
//
// # 1. Registering Routes
//
// Define routes inside the plugin and register them into the lariv.Plugin stages:
//
//	package myplugin
//
//	import (
//		"net/http"
//		"github.com/lariv-in/lariv"
//		"github.com/lariv-in/lariv/registry"
//	)
//
//	func handleDetails(w http.ResponseWriter, r *http.Request) {
//		id := r.PathValue("id")
//		w.Write([]byte("Fetching details for: " + id))
//	}
//
//	func pluginRoutes() lariv.PluginFeatures[lariv.Route] {
//		return lariv.PluginFeatures[lariv.Route]{
//			Entries: []registry.Pair[string, lariv.Route]{
//				{
//					Key: "myplugin.details",
//					Value: lariv.Route{
//						Path:    "/details/{id}/",
//						Handler: http.HandlerFunc(handleDetails),
//					},
//				},
//			},
//		}
//	}
//
// # 2. Patching Routes
//
// You can patch or override routes declared by other plugins to inject custom middleware or change path details:
//
//	func patchExistingRoutes() lariv.PluginFeatures[lariv.Route] {
//		return lariv.PluginFeatures[lariv.Route]{
//			Patches: []registry.Pair[string, func(lariv.Route) lariv.Route]{
//				{
//					Key: "core.HomeRoute",
//					Value: func(existing lariv.Route) lariv.Route {
//						return lariv.Route{
//							Path:    existing.Path,
//							Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//								w.Write([]byte("Patched landing page!"))
//							}),
//						}
//					},
//				},
//			},
//		}
//	}
//
// # Go Router Reference
//
// For native ServeMux path wildcard rules, see standard library [net/http#ServeMux] documentation.
package routes
