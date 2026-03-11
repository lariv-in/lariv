package p_courses

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
)

const AppUrl = "/courses/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugins.Register("p_courses", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "book-open",
		Url:         u,
		VerboseName: "Courses",
	})
	if err != nil {
		log.Panic(err)
	}
}

