package p_teachers

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/teachers/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_teachers", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "academic-cap",
		URL:         u,
		VerboseName: "Teachers",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
