package p_nirmancampus_programs

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_programs"
)

func init() {
	u, err := url.Parse(p_programs.AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_programs", lago.Plugin{
		Type:        lago.PluginTypeAddon,
		Icon:        "academic-cap",
		URL:         u,
		VerboseName: "Programs (Nirmancampus)",
	})
	if err != nil {
		log.Panic(err)
	}
}
