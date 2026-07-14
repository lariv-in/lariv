// Package app contains explanations and code examples for the plugin app.go file in Lago.
//
// # Plugin Definition (app.go)
//
// Every plugin must define an app.go file that implements the Plugin entrypoint.
// This is done by implementing a GetPlugin function returning a registry.Pair wrapping
// the plugin's key and its lago.Plugin configuration struct.
//
// The plugin registers lifecycle stages like Views, Pages, Routes, Models, CommandFactories, and Migrations.
//
// # Example GetPlugin Signature
//
//	package myplugin
//
//	import (
//		"github.com/lariv-in/lago"
//		"github.com/lariv-in/lago/registry"
//	)
//
//	// GetPlugin returns the registry contributions for this plugin.
//	func GetPlugin() registry.Pair[string, lago.Plugin] {
//		return registry.Pair[string, lago.Plugin]{
//			Key: "myplugin",
//			Value: lago.Plugin{
//				Type:             lago.PluginTypeApp, // or lago.PluginTypeAddon
//				VerboseName:      "My Custom Feature",
//				Views:            lago.PluginStages(pluginViews),
//				Pages:            lago.PluginStages(pluginPages),
//				Routes:           lago.PluginStages(pluginRoutes),
//				Models:           lago.PluginStages(pluginModels),
//				Migrations:       lago.PluginStages(pluginMigrations),
//				CommandFactories: lago.PluginStages(pluginCommandFactories),
//			},
//		}
//	}
package app
