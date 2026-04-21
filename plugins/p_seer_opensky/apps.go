package p_seer_opensky

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/seer-opensky/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	err = lago.RegistryPlugin.Register("p_seer_opensky", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "map-pin",
		URL:         u,
		VerboseName: "OpenSky flights",
	})
	if err != nil {
		log.Panic(err)
	}
}
