package p_totschool_tally

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
)

const AppUrl = "/tally/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_totschool_tally", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "chart-bar",
		URL:         u,
		VerboseName: "Progress Tracker",
	})
	if err != nil {
		log.Panic(err)
	}
}
