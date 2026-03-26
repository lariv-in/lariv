package p_nirmancampus_programs

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/programs/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_programs", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "academic-cap",
		URL:         u,
		VerboseName: "Programs",
	})
	if err != nil {
		log.Panic(err)
	}
}
