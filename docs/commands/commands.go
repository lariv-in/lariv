// Package commands contains explanations and code examples for plugin-specific CLI commands in Lago.
//
// # CLI Commands (commands.go)
//
// Plugins can contribute custom CLI commands to the main application's command wrapper.
// This is achieved by creating command factories using the spf13/cobra library.
//
// # Command Factory and Registration Example
//
//	package myplugin
//
//	import (
//		"fmt"
//		"github.com/lariv-in/lago"
//		"github.com/lariv-in/lago/registry"
//		"github.com/spf13/cobra"
//	)
//
//	func createGreetCommand(config lago.LagoConfig) *cobra.Command {
//		cmd := &cobra.Command{
//			Use:   "greet",
//			Short: "Greet a developer",
//			Run: func(cmd *cobra.Command, args []string) {
//				name, _ := cmd.Flags().GetString("name")
//				fmt.Printf("Hello, %s! Debug environment: %v\n", name, config.Debug)
//			},
//		}
//		cmd.Flags().String("name", "Developer", "Name to greet")
//		return cmd
//	}
//
//	func pluginCommands() lago.PluginFeatures[lago.CommandFactory] {
//		return lago.PluginFeatures[lago.CommandFactory]{
//			Entries: []registry.Pair[string, lago.CommandFactory]{
//				{Key: "myplugin.greet", Value: createGreetCommand},
//			},
//		}
//	}
//
//	// Registering inside lago.Plugin
//	lago.Plugin{
//		CommandFactories: lago.PluginStages(pluginCommands),
//	}
//
// # Cobra Reference
//
// For more CLI arguments, flags, and validations details, refer to the [github.com/spf13/cobra] package documentation.
package commands
