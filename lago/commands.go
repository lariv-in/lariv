package lago

import (
	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/lariv-in/lago/registry"
)

func Start(config LagoConfig) error {
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

	rootCmd.AddCommand(&cobra.Command{
		Use:   "tui",
		Short: "Launch the TUI instead of running the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := InitDB(config)
			if err != nil {
				return err
			}
			_, err = tea.NewProgram(initialModel(db)).Run()
			return err
		},
	})

	for _, pair := range *RegistryCommand.AllStable(registry.AlphabeticalByKey[CommandFactory]{}) {
		rootCmd.AddCommand(pair.Value(config))
	}

	return rootCmd.Execute()
}
