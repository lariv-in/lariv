package p_nirmancampus_website

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/website/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	if err := lago.RegistryPlugin.Register("p_nirmancampus_website", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "globe-alt",
		URL:         u,
		VerboseName: "Website",
		Roles:       []string{"nirmancampus_admin"},
	}); err != nil {
		log.Panic(err)
	}
}

