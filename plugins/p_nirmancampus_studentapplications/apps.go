package p_nirmancampus_studentapplications

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/student-applications/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_studentapplications", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "document-text",
		URL:         u,
		VerboseName: "Student applications",
		Roles:       []string{"superuser", "nirmancampus_admin"},
	})
	if err != nil {
		log.Panic(err)
	}
}
