package p_no_signup

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

// GetPlugin returns registry contributions for [lariv.BuildAllRegistries].
func GetPlugin() registry.Pair[string, lariv.Plugin] {
	return registry.Pair[string, lariv.Plugin]{
		Key: "p_no_signup",
		Value: lariv.Plugin{
			Type:        lariv.PluginTypeAddon,
			VerboseName: "No Signup",
			Pages:       lariv.PluginStages(pluginPages),
			Routes:      lariv.PluginStages(pluginRoutes),
		},
	}
}
