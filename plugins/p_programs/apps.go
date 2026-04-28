package p_programs

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/programs/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_programs", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "squares-2x2",
		URL:         u,
		VerboseName: "Programs",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
