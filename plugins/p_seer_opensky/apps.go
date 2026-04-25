package p_seer_opensky

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

// AppUrl is the base path for the OpenSky plugin UI.
const AppUrl = "/seer-opensky/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	err = lago.RegistryPlugin.Register("p_seer_opensky", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "paper-airplane",
		URL:         u,
		VerboseName: "OpenSky",
	})
	if err != nil {
		log.Panic(err)
	}
}
