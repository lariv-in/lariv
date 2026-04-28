package p_attendance

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/attendance/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_attendance", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "clipboard-document-check",
		URL:         u,
		VerboseName: "Attendance",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
