package p_admissions

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/admissions/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_admissions", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "user-plus",
		URL:         u,
		VerboseName: "Admissions",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
