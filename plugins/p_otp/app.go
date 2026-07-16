package p_otp

import (
	"log"
	"net/url"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

const AppURL = "/otp/preferences/"

// GetPlugin returns registry contributions for [lariv.BuildAllRegistries].
func GetPlugin() registry.Pair[string, lariv.Plugin] {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}

	return registry.Pair[string, lariv.Plugin]{
		Key: "p_otp",
		Value: lariv.Plugin{
			Type:        lariv.PluginTypeApp,
			Icon:        "key",
			URL:         u,
			VerboseName: "OTP Preferences",
			Roles:       []string{""},
			Migrations:  pluginStages(pluginMigrations),
			Views:       pluginStages(pluginViews),
			Pages:       pluginStages(pluginPages),
			Routes:      pluginStages(pluginRoutes),
			Models:      pluginStages(pluginModels),
		},
	}
}
