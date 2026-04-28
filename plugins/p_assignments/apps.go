package p_assignments

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/assignments/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_assignments", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "document-text",
		URL:         u,
		VerboseName: "Assignments",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
