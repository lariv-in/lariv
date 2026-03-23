package p_assignments

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/assignments/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_assignments", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "document-text",
		URL:         u,
		VerboseName: "Assignments",
	})
	if err != nil {
		log.Panic(err)
	}
}
