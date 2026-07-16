package lariv

import (
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/registry"
	"github.com/lariv-in/lariv/views"
)

// FillRegistry merges feature bundles from plugins, then populates and assigns an immutable registry
// after executing [PluginFeatures.Build] for each element block. Patches must be pure and idempotent.
func FillRegistry[T any](features [][]func() PluginFeatures[T], targetRegistry *registry.ImmutableRegistry[T]) {
	finalFeatures := PluginFeatures[T]{}
	for _, feature := range features {
		if feature == nil {
			continue
		}
		for _, featureFn := range feature {
			finalFeatures = finalFeatures.Merge(featureFn())
			*targetRegistry = registry.NewImmutableRegistry(finalFeatures.Build())
		}
	}
}

// MapSlice maps elements in a slice from type T to type R using a converter function.
func MapSlice[T any, R any](slice []T, mapper func(T) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = mapper(v)
	}
	return result
}

// BuildAllRegistries executes mapping and populates all application registries using the slice of active plugins.
// It sets up migrations, views, configs, DB hooks, routers, models, layers, and page templates.
//
// Use Cases:
//   - Initializing registries at startup before starting the server.
//
// Example:
//
//	lariv.BuildAllRegistries(allActivePlugins)
func BuildAllRegistries(allPlugins []registry.Pair[string, Plugin]) {
	FillRegistry(MapSlice(allPlugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[UsefulFilesystem] {
		return pair.Value.Migrations
	}), RegistryMigrations)
	FillRegistry(MapSlice(allPlugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[DBInitHook] {
		return pair.Value.DBInitHooks
	}), RegistryDBInit)
	FillRegistry(MapSlice(allPlugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[Config] {
		return pair.Value.Configs
	}), RegistryConfig)
	FillRegistry(MapSlice(allPlugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[Generator] {
		return pair.Value.Generators
	}), RegistryGenerator)
	FillRegistry(MapSlice(allPlugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[CommandFactory] {
		return pair.Value.CommandFactories
	}), RegistryCommand)
	FillRegistry(MapSlice(allPlugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[any] {
		return pair.Value.Models
	}), RegistryModel)
	FillRegistry(MapSlice(allPlugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[views.GlobalLayer] {
		return pair.Value.Layers
	}), RegistryLayer)
	FillRegistry(MapSlice(allPlugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[components.PageInterface] {
		return pair.Value.Pages
	}), RegistryPage)
	FillRegistry(MapSlice(allPlugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[*views.View] {
		return pair.Value.Views
	}), RegistryView)
	FillRegistry(MapSlice(allPlugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[Route] {
		return pair.Value.Routes
	}), RegistryRoute)

	// Installed-plugin metadata for tools like dashboard.AppsGrid (PluginType filter, RBAC tiles).
	*RegistryPlugin = registry.NewImmutableRegistry(allPlugins)
}
