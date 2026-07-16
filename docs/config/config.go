// Package config contains explanations and code examples for plugin configurations in Lariv.
//
// # Plugin Configurations (config.go)
//
// Plugins can define their own configuration structs that map directly from settings in the main config.toml file.
// The configuration struct must implement the lariv.Config interface.
//
// # Example Config Definition
//
//	package myplugin
//
//	import (
//		"github.com/lariv-in/lariv"
//		"github.com/lariv-in/lariv/registry"
//	)
//
//	type MyPluginConfig struct {
//		ApiToken    string `toml:"api_token"`
//		MaxRetries  int    `toml:"max_retries"`
//		EnableCache bool   `toml:"enable_cache"`
//	}
//
//	// PostConfig executes sanity checks and assigns default values after TOML values are loaded.
//	func (c *MyPluginConfig) PostConfig() {
//		if c.MaxRetries <= 0 {
//			c.MaxRetries = 3
//		}
//	}
//
//	var PluginConfig = &MyPluginConfig{}
//
//	func pluginConfigs() lariv.PluginFeatures[lariv.Config] {
//		return lariv.PluginFeatures[lariv.Config]{
//			Entries: []registry.Pair[string, lariv.Config]{
//				{Key: "myplugin", Value: PluginConfig},
//			},
//		}
//	}
//
//	// Registering inside lariv.Plugin
//	lariv.Plugin{
//		Configs: lariv.PluginStages(pluginConfigs),
//	}
package config
