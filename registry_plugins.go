package lago

import (
	"net/http"
	"net/url"
	"slices"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// PluginType represents the classification of a Lago plugin.
type PluginType int

const (
	// PluginTypeApp indicates standalone, self-contained plugins containing primary application business logic and models.
	PluginTypeApp = iota
	// PluginTypeAddon indicates plugins that will not get listed on the dashboard's app grid.
	PluginTypeAddon
	// PluginTypeService indicates plugins booting long-running processes (e.g. background queue workers, SSE connections).
	PluginTypeService
)

// PluginFeatures collects registry Entries and optional Patches for a specific feature type T.
// Patches are applied to registered entries with matching keys during registry builds.
type PluginFeatures[T any] struct {
	// Entries represents the slice of new registry contributions to register.
	Entries []registry.Pair[string, T]
	// Patches represents the slice of modification rules targeting previously registered components with matching keys.
	Patches []registry.Pair[string, func(T) T]
}

// Build returns registry pairs with patches applied in registration order.
// Patches must be pure and idempotent: they must not mutate their argument T in place, and
// applying the same patch again to its own output must yield an equivalent result. Registry
// assembly may invoke Build more than once while merging plugins.
func (f *PluginFeatures[T]) Build() []registry.Pair[string, T] {
	entries := slices.Clone(f.Entries)
	for i := range len(entries) {
		for _, v := range f.Patches {
			if v.Key != entries[i].Key {
				continue
			}
			entries[i].Value = v.Value(entries[i].Value)
		}
	}
	return entries
}

// Merge concatenates Entries and Patches. Order is preserved so patch application order stays deterministic.
func (f PluginFeatures[T]) Merge(others ...PluginFeatures[T]) PluginFeatures[T] {
	result := f
	for _, other := range others {
		result.Entries = append(result.Entries, other.Entries...)
		result.Patches = append(result.Patches, other.Patches...)
	}
	return result
}

// Plugin defines a collection of application features, CLI commands, database initializers, routing endpoints, configurations, and metadata components.
//
// Use Cases:
//   - Defining modular code boundaries to group related database tables, templates, middlewares, and views.
//
// Example Definition:
//
//	var DashboardPlugin = lago.Plugin{
//		Type:        lago.PluginTypeApp,
//		VerboseName: "Dashboard",
//		Pages: lago.PluginStages(func() lago.PluginFeatures[components.PageInterface] {
//			return lago.PluginFeatures[components.PageInterface]{
//				Entries: []registry.Pair[string, components.PageInterface]{
//					registry.NewPair("dashboard.home", DashboardHome{}),
//				},
//			}
//		}),
//	}
type Plugin struct {
	// Type specifies the plugin classification.
	Type PluginType
	// Icon represents a CSS class or emoji string for UI representation.
	Icon string
	// URL represents the primary entry URL path pointing to the plugin landing page.
	URL *url.URL
	// VerboseName represents the user-friendly human readable label for the plugin.
	VerboseName string
	// Roles lists the authorized security roles allowed to interact with the plugin.
	Roles []string
	// Migrations defines embedded database schema update folders.
	Migrations []func() PluginFeatures[UsefulFilesystem]
	// Views defines views controllers to map request pipelines.
	Views []func() PluginFeatures[*views.View]
	// Routes maps endpoint strings to standard HTTP handlers.
	Routes []func() PluginFeatures[Route]
	// Pages registers HTML component page layout trees.
	Pages []func() PluginFeatures[components.PageInterface]
	// Models registers GORM schema models.
	Models []func() PluginFeatures[any]
	// Layers registers global request/response multiplexer middlewares.
	Layers []func() PluginFeatures[views.GlobalLayer]
	// Generators registers database mock seed handlers.
	Generators []func() PluginFeatures[Generator]
	// DBInitHooks registers GORM initialization hook decorators.
	DBInitHooks []func() PluginFeatures[DBInitHook]
	// Configs registers custom setting configurations structs.
	Configs []func() PluginFeatures[Config]
	// CommandFactories registers custom CLI commands generators.
	CommandFactories []func() PluginFeatures[CommandFactory]
}

// PluginStages wraps a single feature callback as a one-element slice for [Plugin] fields.
func PluginStages[T any](stage func() PluginFeatures[T]) []func() PluginFeatures[T] {
	return []func() PluginFeatures[T]{stage}
}

// RegistryPlugin represents the global immutable registry tracking installed plugins metadata.
var RegistryPlugin *registry.ImmutableRegistry[Plugin] = &registry.ImmutableRegistry[Plugin]{}

// CorePlugin creates the framework core plugin configuration.
// It registers standard global middleware layers (e.g. database attachments, logging, caching controls, environment mappings)
// and hooks up default routing and landing page views.
func CorePlugin(db *gorm.DB, config LagoConfig) registry.Pair[string, Plugin] {
	layers := PluginFeatures[views.GlobalLayer]{}
	layers.Entries = append(layers.Entries, registry.Pair[string, views.GlobalLayer]{Key: "core.AttachRequestLayer", Value: views.AttachRequestLayer{}})
	layers.Entries = append(layers.Entries, registry.Pair[string, views.GlobalLayer]{Key: "core.DbLayer", Value: DBLayer{DB: db}})
	if config.Debug {
		layers.Entries = append(layers.Entries, registry.Pair[string, views.GlobalLayer]{Key: "core.LoggingLayer", Value: LoggingLayer{}})
		layers.Entries = append(layers.Entries, registry.Pair[string, views.GlobalLayer]{Key: "core.CacheDisableLayer", Value: CacheDisableLayer{}})
	}
	layers.Entries = append(layers.Entries, registry.Pair[string, views.GlobalLayer]{Key: "core.HtmxBoostLayer", Value: HtmxBoostLayer{}})
	layers.Entries = append(layers.Entries, registry.Pair[string, views.GlobalLayer]{Key: "core.EnvironmentLayer", Value: EnvironmentLayer{}})

	return registry.Pair[string, Plugin]{
		Key: "core", Value: Plugin{
			Type: PluginTypeAddon,
			URL: &url.URL{
				Path: "/",
			},
			VerboseName: "Core",
			Roles:       []string{"superuser", "admin"},
			Views: []func() PluginFeatures[*views.View]{
				func() PluginFeatures[*views.View] {
					return PluginFeatures[*views.View]{
						Entries: []registry.Pair[string, *views.View]{
							{Key: "core.HomeView", Value: GetPageView("core.HomePage")},
						},
					}
				},
			},
			Pages: []func() PluginFeatures[components.PageInterface]{
				func() PluginFeatures[components.PageInterface] {
					return PluginFeatures[components.PageInterface]{
						Entries: []registry.Pair[string, components.PageInterface]{
							{Key: "core.HomePage", Value: components.ShellBase{}},
						},
					}
				},
			},
			Layers: []func() PluginFeatures[views.GlobalLayer]{
				func() PluginFeatures[views.GlobalLayer] {
					return layers
				},
			},
			Routes: []func() PluginFeatures[Route]{
				func() PluginFeatures[Route] {
					return PluginFeatures[Route]{
						Entries: []registry.Pair[string, Route]{
							{Key: "core.HomeRoute", Value: Route{Path: "/", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								w.WriteHeader(http.StatusOK)
								w.Write([]byte("Hello, World!"))
							})}},
						},
					}
				},
			},
		},
	}
}
