package p_assignments_semesters

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_assignments"
)

func init() {
	u, err := url.Parse(p_assignments.AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_assignments_semesters", lago.Plugin{
		Type:        lago.PluginTypeAddon,
		Icon:        "calendar",
		URL:         u,
		VerboseName: "Assignments (Semesters)",
	})
	if err != nil {
		log.Panic(err)
	}
}
