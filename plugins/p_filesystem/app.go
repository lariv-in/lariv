package p_filesystem

import (
	"log"
	"net/url"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

const AppUrl = "/filesystem/"

// GetPlugin returns the registry contributions for this plugin for [lariv.BuildAllRegistries].
// Callers assembling the full plugin list should include a pair with key "p_filesystem" and this value.
func GetPlugin() registry.Pair[string, lariv.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	return registry.Pair[string, lariv.Plugin]{
		Key: "p_filesystem", Value: lariv.Plugin{
			Type:        lariv.PluginTypeApp,
			Icon:        "folder",
			URL:         u,
			VerboseName: "Filesystem",
			Roles:       []string{"superuser", "admin"},
			Migrations:  pluginStages(pluginMigrations),
			Views:       pluginStages(pluginViews),
			Pages:       pluginStages(pluginPages),
			Routes:      pluginStages(pluginRoutes),
			Configs:     pluginStages(pluginConfigs),
			Generators:  pluginStages(pluginGenerators),
		},
	}
}
