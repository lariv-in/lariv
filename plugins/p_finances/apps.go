package p_finances

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/finances/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_finances", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "currency-dollar",
		URL:         u,
		VerboseName: "Finances",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
