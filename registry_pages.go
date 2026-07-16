package lariv

import (
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/registry"
)

// RegistryPage represents the global immutable registry tracking page templates of type [components.PageInterface].
// Pages registered here are resolved by view controllers to compile dynamic user interfaces.
//
// Use Cases:
//   - Bundling plugin views structures (e.g. login pages, settings layout templates).
//
// Example Definition:
//
//	type HelloPage struct {
//		components.Page
//	}
//
//	func (p HelloPage) Build(ctx context.Context) gomponents.Node {
//		return html.Div(html.H1(gomponents.Text("Hello World")))
//	}
//
// Example Registration:
//
//	// In your lariv.Plugin setup:
//	lariv.Plugin{
//		Pages: lariv.PluginStages(func() PluginFeatures[components.PageInterface] {
//			return PluginFeatures[components.PageInterface]{
//				Entries: []registry.Pair[string, components.PageInterface]{
//					registry.NewPair("hello_page", HelloPage{}),
//				},
//			}
//		}),
//	}
//
// Example Patch:
//
//	// Register a patch to decorate or modify pages from another plugin:
//	lariv.Plugin{
//		Pages: lariv.PluginStages(func() PluginFeatures[components.PageInterface] {
//			return PluginFeatures[components.PageInterface]{
//				Patches: []registry.Pair[string, func(components.PageInterface) components.PageInterface]{
//					registry.NewPair("hello_page", func(existing components.PageInterface) components.PageInterface {
//						// Modify layout or add wrapper children:
//						return existing
//					}),
//				},
//			}
//		}),
//	}
//
// Example Retrieval:
//
//	page, ok := RegistryPage.Get("hello_page")
var RegistryPage *registry.ImmutableRegistry[components.PageInterface] = &registry.ImmutableRegistry[components.PageInterface]{}
