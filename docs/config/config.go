// Package config contains explanations and code examples for plugin configurations in Lago.
//
// # Plugin Configurations (config.go)
//
// Plugins can define their own configuration structs that map directly from settings in the main config.toml file.
// The configuration struct must implement the lago.Config interface.
//
// # Example Config Definition
//
//	package myplugin
//
//	import (
//		"github.com/lariv-in/lago"
//		"github.com/lariv-in/lago/registry"
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
//	func pluginConfigs() lago.PluginFeatures[lago.Config] {
//		return lago.PluginFeatures[lago.Config]{
//			Entries: []registry.Pair[string, lago.Config]{
//				{Key: "myplugin", Value: PluginConfig},
//			},
//		}
//	}
//
//	// Registering inside lago.Plugin
//	lago.Plugin{
//		Configs: lago.PluginStages(pluginConfigs),
//	}
package config
