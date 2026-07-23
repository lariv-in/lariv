package p_blog

import (
	"net/url"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

const AppURL = "/blog/"

// GetPlugin returns registry contributions for [lariv.BuildAllRegistries].
func GetPlugin() registry.Pair[string, lariv.Plugin] {
	u, _ := url.Parse(AppURL)

	return registry.Pair[string, lariv.Plugin]{
		Key: "p_blog",
		Value: lariv.Plugin{
			Type:        lariv.PluginTypeApp,
			Icon:        "newspaper",
			URL:         u,
			VerboseName: "Blog",
			Roles:       []string{"superuser", "admin"},
			Migrations:  lariv.PluginStages(pluginMigrations),
			Models:      lariv.PluginStages(pluginModels),
			Views:       lariv.PluginStages(pluginViews),
			Pages:       lariv.PluginStages(pluginPages),
			Routes:      lariv.PluginStages(pluginRoutes),
		},
	}
}
