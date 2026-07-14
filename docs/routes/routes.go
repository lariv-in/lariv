// Package routes contains explanations and code examples for HTTP routing in Lago.
//
// # HTTP Routing (routes.go)
//
// Lago routes map request path patterns directly to handlers.
// The routing layer is built on top of Go 1.22+'s native net/http router (ServeMux), supporting path variables natively.
//
// # 1. Registering Routes
//
// Define routes inside the plugin and register them into the lago.Plugin stages:
//
//	package myplugin
//
//	import (
//		"net/http"
//		"github.com/lariv-in/lago"
//		"github.com/lariv-in/lago/registry"
//	)
//
//	func handleDetails(w http.ResponseWriter, r *http.Request) {
//		id := r.PathValue("id")
//		w.Write([]byte("Fetching details for: " + id))
//	}
//
//	func pluginRoutes() lago.PluginFeatures[lago.Route] {
//		return lago.PluginFeatures[lago.Route]{
//			Entries: []registry.Pair[string, lago.Route]{
//				{
//					Key: "myplugin.details",
//					Value: lago.Route{
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
//	func patchExistingRoutes() lago.PluginFeatures[lago.Route] {
//		return lago.PluginFeatures[lago.Route]{
//			Patches: []registry.Pair[string, func(lago.Route) lago.Route]{
//				{
//					Key: "core.HomeRoute",
//					Value: func(existing lago.Route) lago.Route {
//						return lago.Route{
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
