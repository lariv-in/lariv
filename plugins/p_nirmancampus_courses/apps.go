package p_nirmancampus_courses

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/courses/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_courses", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "book-open",
		URL:         u,
		VerboseName: "Courses",
		Roles:       []string{"student", "admin"},
	})
	if err != nil {
		log.Panic(err)
	}
}

