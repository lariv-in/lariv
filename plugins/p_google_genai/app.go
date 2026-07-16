package p_google_genai

import (
	"log"
	"net/url"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

const AppUrl = "/google-genai/"

// GetPlugin returns registry contributions for [lariv.BuildAllRegistries].
func GetPlugin() registry.Pair[string, lariv.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	return registry.Pair[string, lariv.Plugin]{
		Key: "p_google_genai",
		Value: lariv.Plugin{
			// Addon: not listed on dashboard Apps grid; API key consumed by other plugins.
			Type:        lariv.PluginTypeAddon,
			Icon:        "sparkles",
			URL:         u,
			VerboseName: "Google GenAI",
			Configs:     pluginStages(pluginConfigs),
		},
	}
}
