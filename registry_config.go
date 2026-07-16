package lariv

import (
	"github.com/lariv-in/lariv/registry"
)

// Config defines the interface implemented by plugin config structs to receive and validate parsed settings from TOML files.
// PostConfig is executed automatically after settings are mapped, enabling validation or setting default values.
//
// Use Cases:
//   - Defining configuration tables for plugins (e.g. storage paths, API client secrets).
//
// Example Definition:
//
//	type DashboardConfig struct {
//		AppName string
//	}
//
//	func (c *DashboardConfig) PostConfig() {
//		if c.AppName == "" {
//			c.AppName = "My Dashboard App"
//		}
//	}
//
// Example Registration:
//
//	var DashboardConfigPtr = &DashboardConfig{}
//
//	// Register the config instance inside your lariv.Plugin configuration:
//	lariv.Plugin{
//		Configs: lariv.PluginStages(func() PluginFeatures[Config] {
//			return PluginFeatures[Config]{
//				Entries: []registry.Pair[string, Config]{
//					registry.NewPair("dashboard", DashboardConfigPtr),
//				},
//			}
//		}),
//	}
//
// Example Patch:
//
//	// Register a patch to modify config settings from another plugin:
//	lariv.Plugin{
//		Configs: lariv.PluginStages(func() PluginFeatures[Config] {
//			return PluginFeatures[Config]{
//				Patches: []registry.Pair[string, func(Config) Config]{
//					registry.NewPair("dashboard", func(existing Config) Config {
//						cfg := existing.(*DashboardConfig)
//						cfg.AppName = "Modified App Name"
//						return cfg
//					}),
//				},
//			}
//		}),
//	}
//
// Example Retrieval:
//
//	cfgVal, ok := RegistryConfig.Get("dashboard")
type Config interface {
	// PostConfig executes sanity checks and assigns default values after TOML values are loaded.
	PostConfig()
}

// RegistryConfig represents the global immutable registry mapping config identifiers to their Config instances.
var RegistryConfig *registry.ImmutableRegistry[Config] = &registry.ImmutableRegistry[Config]{}
