package p_seer_intel

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/seer-intel/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_seer_intel", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "eye",
		URL:         u,
		VerboseName: "Intel",
	})
	if err != nil {
		log.Panic(err)
	}
}
