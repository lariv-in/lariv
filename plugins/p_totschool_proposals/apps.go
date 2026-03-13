package p_totschool_proposals

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
)

const AppUrl = "/proposals/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	err = lago.RegistryPlugins.Register("p_totschool_proposals", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "document-text",
		URL:         u,
		VerboseName: "Proposals",
	})
	if err != nil {
		log.Panic(err)
	}
}
