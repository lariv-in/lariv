package p_allocation

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/allocation/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_allocation", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "rectangle-group",
		URL:         u,
		VerboseName: "Allocation",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
