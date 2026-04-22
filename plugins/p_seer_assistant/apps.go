package p_seer_assistant

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

// AppUrl is the HTTP prefix for this plugin (trailing slash).
const AppUrl = "/seer-assistant/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	err = lago.RegistryPlugin.Register("p_seer_assistant", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "sparkles",
		URL:         u,
		VerboseName: "Assistant",
	})
	if err != nil {
		log.Panic(err)
	}
}
