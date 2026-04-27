package p_semesters

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/semesters/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_semesters", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "calendar-days",
		URL:         u,
		VerboseName: "Semesters",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
