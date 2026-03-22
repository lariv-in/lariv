package p_courses_teachers

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_courses"
)

func init() {
	u, err := url.Parse(p_courses.AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_courses_teachers", lago.Plugin{
		Type:        lago.PluginTypeAddon,
		Icon:        "link",
		URL:         u,
		VerboseName: "Courses Teachers",
	})
	if err != nil {
		log.Panic(err)
	}
}
