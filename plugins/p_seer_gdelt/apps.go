package p_seer_gdelt

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/seer-gdelt/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	err = lago.RegistryPlugin.Register("p_seer_gdelt", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "magnifying-glass-circle",
		URL:         u,
		VerboseName: "GDELT search",
	})
	if err != nil {
		log.Panic(err)
	}
}
