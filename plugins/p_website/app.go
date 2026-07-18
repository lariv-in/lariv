package p_website

import (
	"log"
	"net/url"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

const AppURL = "/website/"

// GetPlugin returns registry contributions for [lariv.BuildAllRegistries].
func GetPlugin() registry.Pair[string, lariv.Plugin] {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}

	return registry.Pair[string, lariv.Plugin]{
		Key: "p_website",
		Value: lariv.Plugin{
			Type:        lariv.PluginTypeApp,
			Icon:        "globe-alt",
			URL:         u,
			VerboseName: "Website",
			Roles:       []string{"superuser", "admin"},
			Migrations:  pluginStages(pluginMigrations),
			Models:      pluginStages(pluginModels),
			Views:       pluginStages(pluginViews),
			Pages:       pluginStages(pluginPages),
			Routes:      pluginStages(pluginRoutes),
		},
	}
}
