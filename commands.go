package lago

import (
	"github.com/lariv-in/lago/registry"
	"github.com/spf13/cobra"
)

// Start initializes and executes the Cobra CLI application, acting as the main entrypoint for any Lago application.
//
// CLI Command Scopes:
//   - Root Command: Starts the HTTP web server via [StartServer].
//   - generate: Runs database seed generators via [RunGenerators].
//   - tui: Launches the Bubble Tea terminal user interface.
//   - Plugin Commands: Resolves and registers custom commands dynamically loaded from [RegistryCommand].
//
// Registries and configurations must be populated before invoking this function (e.g. using LoadConfigFromFile).
//
// Use Cases:
//   - Initializing the CLI bootstrapper in the main execution block of a Go application.
//
// Example:
//
//	func main() {
//		config := lago.LagoConfig{
//			DBType:  lago.DBTypePostgres,
//			Address: ":8080",
//		}
//		plugins := []registry.Pair[string, lago.Plugin]{
//			p_dashboard.GetPlugin(),
//		}
//		if err := lago.Start(config, plugins); err != nil {
//			log.Fatal(err)
//		}
//	}
func Start(config LagoConfig, plugins []registry.Pair[string, Plugin]) error {
	_ = plugins
	rootCmd := &cobra.Command{
		Use:   "lago",
		Short: "Lago web framework",
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartServer(config)
		},
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "generate",
		Short: "Run data generators to seed the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			RunGenerators(config)
			return nil
		},
	})

	for _, pair := range *RegistryCommand.AllStable() {
		rootCmd.AddCommand(pair.Value(config))
	}

	return rootCmd.Execute()
}
