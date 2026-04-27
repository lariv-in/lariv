package p_announcements

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/announcements/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_announcements", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "megaphone",
		URL:         u,
		VerboseName: "Announcements",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
