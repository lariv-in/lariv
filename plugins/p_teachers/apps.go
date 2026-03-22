package p_teachers

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/teachers/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_teachers", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "academic-cap",
		URL:         u,
		VerboseName: "Teachers",
	})
	if err != nil {
		log.Panic(err)
	}
}
