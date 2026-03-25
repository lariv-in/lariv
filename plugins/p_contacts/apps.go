package p_contacts

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/contacts/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_contacts", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "user-group",
		URL:         u,
		VerboseName: "Contacts",
	})
	if err != nil {
		log.Panic(err)
	}
}
