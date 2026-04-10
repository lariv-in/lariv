package lago

import (
	"github.com/lariv-in/lago/registry"
	"github.com/spf13/cobra"
)

type CommandFactory func(LagoConfig) *cobra.Command

var RegistryCommand *registry.Registry[CommandFactory] = registry.NewRegistry[CommandFactory]()
