package p_google_genai

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/google-genai/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_google_genai", lago.Plugin{
		Type:        lago.PluginTypeAddon, // not listed on dashboard Apps grid; routes/status UI still registered
		Icon:        "sparkles",
		URL:         u,
		VerboseName: "Google GenAI",
	}); err != nil {
		log.Panic(err)
	}
}
