package p_courses

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/courses/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_courses", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "book-open",
		URL:         u,
		VerboseName: "Courses",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
