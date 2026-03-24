package p_announcements_semesters

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_announcements"
)

func init() {
	u, err := url.Parse(p_announcements.AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_announcements_semesters", lago.Plugin{
		Type:        lago.PluginTypeAddon,
		Icon:        "calendar",
		URL:         u,
		VerboseName: "Announcements (Semesters)",
	})
	if err != nil {
		log.Panic(err)
	}
}
