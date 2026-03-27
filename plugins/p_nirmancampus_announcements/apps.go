package p_nirmancampus_announcements

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/announcements/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_announcements", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "megaphone",
		URL:         u,
		VerboseName: "Announcements",
		Roles:       []string{"superuser", "admin", "student"},
	})
	if err != nil {
		log.Panic(err)
	}
}
