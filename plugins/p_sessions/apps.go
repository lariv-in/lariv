package p_sessions

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/sessions/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_sessions", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "clock",
		URL:         u,
		VerboseName: "Sessions",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
