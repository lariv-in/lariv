package lariv

import (
	"github.com/lariv-in/lariv/registry"
	"github.com/lariv-in/lariv/views"
)

// RegistryLayer represents the global immutable registry tracking global middleware layers of type [views.GlobalLayer].
// Global layers act as multiplexer-level middlewares wrapped around request paths prior to route matching.
//
// Use Cases:
//   - Registering global HTTP request filters (e.g. CORS headers, request loggers, CSRF validation, compression).
//
// Example Definition:
//
//	type CorsLayer struct{}
//
//	func (l CorsLayer) Next(next http.Handler) http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			w.Header().Set("Access-Control-Allow-Origin", "*")
//			next.ServeHTTP(w, r)
//		})
//	}
//
// Example Registration:
//
//	// In your lariv.Plugin setup:
//	lariv.Plugin{
//		Layers: lariv.PluginStages(func() PluginFeatures[views.GlobalLayer] {
//			return PluginFeatures[views.GlobalLayer]{
//				Entries: []registry.Pair[string, views.GlobalLayer]{
//					registry.NewPair("cors_layer", CorsLayer{}),
//				},
//			}
//		}),
//	}
//
// Example Patch:
//
//	// Register a patch to replace or modify global layers from another plugin:
//	lariv.Plugin{
//		Layers: lariv.PluginStages(func() PluginFeatures[views.GlobalLayer] {
//			return PluginFeatures[views.GlobalLayer]{
//				Patches: []registry.Pair[string, func(views.GlobalLayer) views.GlobalLayer]{
//					registry.NewPair("cors_layer", func(existing views.GlobalLayer) views.GlobalLayer {
//						// Wrap or extend CorsLayer:
//						return existing
//					}),
//				},
//			}
//		}),
//	}
//
// Example Retrieval:
//
//	layer, ok := RegistryLayer.Get("cors_layer")
var RegistryLayer *registry.ImmutableRegistry[views.GlobalLayer] = &registry.ImmutableRegistry[views.GlobalLayer]{}
