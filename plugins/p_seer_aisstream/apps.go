package p_seer_aisstream

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/seer-aisstream/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	err = lago.RegistryPlugin.Register("p_seer_aisstream", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "map-pin",
		URL:         u,
		VerboseName: "AISStream ships",
	})
	if err != nil {
		log.Panic(err)
	}
}
