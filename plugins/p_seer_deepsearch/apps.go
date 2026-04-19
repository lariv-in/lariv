package p_seer_deepsearch

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

// AppUrl is the HTTP prefix for this plugin (trailing slash).
const AppUrl = "/seer-deepsearch/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	err = lago.RegistryPlugin.Register("p_seer_deepsearch", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "magnifying-glass-circle",
		URL:         u,
		VerboseName: "Deep search",
	})
	if err != nil {
		log.Panic(err)
	}
}
