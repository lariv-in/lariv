package p_pwa

import (
	"log"
	"net/url"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

// GetPlugin returns registry contributions for [lariv.AppBuilder].
func GetPlugin() registry.Pair[string, lariv.Plugin] {
	u, err := url.Parse("/")
	if err != nil {
		log.Panic(err)
	}
	return registry.Pair[string, lariv.Plugin]{
		Key: "p_pwa",
		Value: lariv.Plugin{
			Type:        lariv.PluginTypeAddon,
			Icon:        "cpu-chip",
			URL:         u,
			VerboseName: "PWA",
			Configs:     pluginStages(pluginConfigs),
			Views:       pluginStages(pluginViews),
			Routes:      pluginStages(pluginRoutes),
			HeadNodes:   lariv.PluginStages(pluginHeadNodes),
		},
	}
}
