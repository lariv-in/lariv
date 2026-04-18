package p_seer_runners

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/seer-runners/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_seer_runners", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "arrow-path",
		URL:         u,
		VerboseName: "Runners",
	})
	if err != nil {
		log.Panic(err)
	}
}
