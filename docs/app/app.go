// Package app contains explanations and code examples for the plugin app.go file in Lariv.
//
// # Plugin Definition (app.go)
//
// Every plugin must define an app.go file that implements the Plugin entrypoint.
// This is done by implementing a GetPlugin function returning a registry.Pair wrapping
// the plugin's key and its lariv.Plugin configuration struct.
//
// The plugin registers lifecycle stages like Views, Pages, Routes, Models, CommandFactories, and Migrations.
//
// # Example GetPlugin Signature
//
//	package myplugin
//
//	import (
//		"github.com/lariv-in/lariv"
//		"github.com/lariv-in/lariv/registry"
//	)
//
//	// GetPlugin returns the registry contributions for this plugin.
//	func GetPlugin() registry.Pair[string, lariv.Plugin] {
//		return registry.Pair[string, lariv.Plugin]{
//			Key: "myplugin",
//			Value: lariv.Plugin{
//				Type:             lariv.PluginTypeApp, // or lariv.PluginTypeAddon
//				VerboseName:      "My Custom Feature",
//				Views:            lariv.PluginStages(pluginViews),
//				Pages:            lariv.PluginStages(pluginPages),
//				Routes:           lariv.PluginStages(pluginRoutes),
//				Models:           lariv.PluginStages(pluginModels),
//				Migrations:       lariv.PluginStages(pluginMigrations),
//				CommandFactories: lariv.PluginStages(pluginCommandFactories),
//			},
//		}
//	}
package app
