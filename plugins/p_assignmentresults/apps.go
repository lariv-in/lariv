package p_assignmentresults

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

	err = lago.RegistryPlugin.Register("p_assignmentresults", lago.Plugin{
		Type:        lago.PluginTypeAddon,
		Icon:        "chart-bar",
		URL:         u,
		VerboseName: "Assignment Results",
	})
	if err != nil {
		log.Panic(err)
	}
}
