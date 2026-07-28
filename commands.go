package lariv

import (
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
type CommandFactory func(LarivConfig) *cobra.Command
