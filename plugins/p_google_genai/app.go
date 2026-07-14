package p_google_genai

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"
)

const AppUrl = "/google-genai/"

// GetPlugin returns registry contributions for [lago.BuildAllRegistries].
func GetPlugin() registry.Pair[string, lago.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	return registry.Pair[string, lago.Plugin]{
		Key: "p_google_genai",
		Value: lago.Plugin{
			// Addon: not listed on dashboard Apps grid; API key consumed by other plugins.
			Type:        lago.PluginTypeAddon,
			Icon:        "sparkles",
			URL:         u,
			VerboseName: "Google GenAI",
			Configs:     pluginStages(pluginConfigs),
		},
	}
}
