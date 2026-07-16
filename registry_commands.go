package lariv

import (
	"github.com/lariv-in/lariv/registry"
	"github.com/spf13/cobra"
)

// CommandFactory represents a generator function that builds Cobra CLI commands mapped to a specific [LarivConfig].
//
// Use Cases:
//   - Defining custom CLI sub-commands inside application plugins (e.g., system diagnostics, database cleaner tasks).
//
// Example:
//
//	var BackupCmdFactory CommandFactory = func(config LarivConfig) *cobra.Command {
//		return &cobra.Command{
//			Use:   "backup",
//			Short: "Executes a database schema backup",
//			Run: func(cmd *cobra.Command, args []string) {
//				executeBackup(config)
//			},
//		}
//	}
//
//	// Register the command factory inside your lariv.Plugin configuration:
//	lariv.Plugin{
//		CommandFactories: lariv.PluginStages(func() PluginFeatures[CommandFactory] {
//			return PluginFeatures[CommandFactory]{
//				Entries: []registry.Pair[string, CommandFactory]{
//					registry.NewPair("backup_db", BackupCmdFactory),
//				},
//			}
//		}),
//	}
//
//	// Register a patch to modify an existing command in another plugin:
//	lariv.Plugin{
//		CommandFactories: lariv.PluginStages(func() PluginFeatures[CommandFactory] {
//			return PluginFeatures[CommandFactory]{
//				Patches: []registry.Pair[string, func(CommandFactory) CommandFactory]{
//					registry.NewPair("backup_db", func(existing CommandFactory) CommandFactory {
//						return func(config LarivConfig) *cobra.Command {
//							cmd := existing(config)
//							cmd.Short = "Patched: " + cmd.Short
//							return cmd
//						}
//					}),
//				},
//			}
//		}),
//	}
//
//	// Retrieve a registered command factory:
//	factory, ok := RegistryCommand.Get("backup_db")
type CommandFactory func(LarivConfig) *cobra.Command

// RegistryCommand represents the global immutable registry mapping custom plugin sub-commands to their CommandFactory builders.
var RegistryCommand *registry.ImmutableRegistry[CommandFactory] = &registry.ImmutableRegistry[CommandFactory]{}
