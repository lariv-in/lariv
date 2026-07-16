package p_users

import (
	"log"
	"net/url"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

// GetPlugin returns the registry contributions for this plugin for [lariv.BuildAllRegistries].
func GetPlugin() registry.Pair[string, lariv.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	return registry.Pair[string, lariv.Plugin]{
		Key: "p_users",
		Value: lariv.Plugin{
			Type:             lariv.PluginTypeApp,
			Icon:             "users",
			URL:              u,
			VerboseName:      "Users",
			Migrations:       pluginStages(pluginMigrations),
			Views:            pluginStages(pluginViews),
			Pages:            pluginStages(pluginPages),
			Routes:           pluginStages(pluginRoutes),
			Models:           pluginStages(pluginModels),
			Generators:       pluginStages(pluginGenerators),
			DBInitHooks:      pluginStages(pluginDBInitHooks),
			Configs:          pluginStages(pluginAuthConfigs),
			CommandFactories: pluginStages(pluginCommandFactories),
		},
	}
}
