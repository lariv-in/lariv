package p_nirmancampus_students

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/students/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_students", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "user",
		URL:         u,
		VerboseName: "Students",
		Roles:       []string{"superuser", "admin", "student"},
	})
	if err != nil {
		log.Panic(err)
	}
}
