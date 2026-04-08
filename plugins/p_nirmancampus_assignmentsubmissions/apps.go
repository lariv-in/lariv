package p_nirmancampus_assignmentsubmissions

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/assignmentsubmissions/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_assignmentsubmissions", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "document-check",
		URL:         u,
		VerboseName: "Assignment Submissions",
		Roles:       []string{"superuser", "admin", "student"},
	})
	if err != nil {
		log.Panic(err)
	}
}
