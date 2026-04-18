package p_seer_reddit

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/seer-reddit/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_seer_reddit", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "chat-bubble-left-right",
		URL:         u,
		VerboseName: "Reddit",
	})
	if err != nil {
		log.Panic(err)
	}
}
