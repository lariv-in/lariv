package p_nirmancampus_courses

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

// AppUrl is under the Programs app; keep in sync with p_nirmancampus_programs.AppUrl + "addon/courses/"
// (courses cannot import programs — import cycle — so the prefix is spelled out here).
var AppUrl = "/programs/addon/courses/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_courses", lago.Plugin{
		Type:        lago.PluginTypeAddon,
		Icon:        "book-open",
		URL:         u,
		VerboseName: "Courses",
		Roles:       []string{"student", "admin"},
	})
	if err != nil {
		log.Panic(err)
	}
}
