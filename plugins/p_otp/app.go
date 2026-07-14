package p_otp

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"
)

const AppURL = "/otp/preferences/"

// GetPlugin returns registry contributions for [lago.BuildAllRegistries].
func GetPlugin() registry.Pair[string, lago.Plugin] {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}

	return registry.Pair[string, lago.Plugin]{
		Key: "p_otp",
		Value: lago.Plugin{
			Type:        lago.PluginTypeApp,
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
