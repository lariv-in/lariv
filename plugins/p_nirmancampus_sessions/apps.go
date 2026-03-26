package p_nirmancampus_sessions

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/semesters/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_sessions", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "calendar",
		URL:         u,
		VerboseName: "Semesters",
	})
	if err != nil {
		log.Panic(err)
	}
}

