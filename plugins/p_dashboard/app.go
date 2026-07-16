package p_dashboard

import (
	"log"
	"net/url"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

const AppUrl = "/dashboard/"

// GetPlugin returns the registry contributions for this plugin (views, pages, routes) for
// [lariv.BuildAllRegistries]. Callers that assemble the full plugin list should include
// a pair with key "p_dashboard" and this value.
func GetPlugin() registry.Pair[string, lariv.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	return registry.Pair[string, lariv.Plugin]{
		Key: "p_dashboard", Value: lariv.Plugin{
			Type:        lariv.PluginTypeAddon,
			Icon:        "dashboard",
			URL:         u,
			VerboseName: "Dashboard",
			Views:       pluginStages(pluginViews),
			Pages:       pluginStages(pluginPages),
			Routes:      pluginStages(pluginRoutes),
		},
	}
}
