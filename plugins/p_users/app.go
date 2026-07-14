package p_users

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"
)

// GetPlugin returns the registry contributions for this plugin for [lago.BuildAllRegistries].
func GetPlugin() registry.Pair[string, lago.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	return registry.Pair[string, lago.Plugin]{
		Key: "p_users",
		Value: lago.Plugin{
			Type:             lago.PluginTypeApp,
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
