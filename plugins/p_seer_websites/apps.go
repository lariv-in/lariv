package p_seer_websites

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

// AppUrl is the HTTP prefix for this plugin (trailing slash).
const AppUrl = "/seer-websites/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	err = lago.RegistryPlugin.Register("p_seer_websites", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "globe-alt",
		URL:         u,
		VerboseName: "Websites",
	})
	if err != nil {
		log.Panic(err)
	}
}
