package lariv

import (
	"github.com/lariv-in/lariv/registry"
)

// RegistryModel represents the global immutable registry tracking application database models of type [any].
// Models registered here are mapped to database structures for schema introspection or auto-migrations.
//
// Use Cases:
//   - Bundling plugin GORM structures (e.g. User, Product) to allow table creation, relationship tracking, or admin panel listing.
//
// Example Definition:
//
//	type Product struct {
//		gorm.Model
//		Name string
//	}
//
// Example Registration:
//
//	// In your lariv.Plugin setup:
//	lariv.Plugin{
//		Models: lariv.PluginStages(func() PluginFeatures[any] {
//			return PluginFeatures[any]{
//				Entries: []registry.Pair[string, any]{
//					registry.NewPair("product_model", Product{}),
//				},
//			}
//		}),
//	}
//
// Example Patch:
//
//	// Register a patch to extend or modify registered models from another plugin:
//	lariv.Plugin{
//		Models: lariv.PluginStages(func() PluginFeatures[any] {
//			return PluginFeatures[any]{
//				Patches: []registry.Pair[string, func(any) any]{
//					registry.NewPair("product_model", func(existing any) any {
//						// Modify or wraps metadata:
//						return existing
//					}),
//				},
//			}
//		}),
//	}
//
// Example Retrieval:
//
//	modelVal, ok := RegistryModel.Get("product_model")
var RegistryModel *registry.ImmutableRegistry[any] = &registry.ImmutableRegistry[any]{}
