package p_nirmancampus_student_zone

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/student-zone/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_student_zone", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "academic-cap",
		URL:         u,
		VerboseName: "Student Zone",
		Roles:       []string{"nirmancampus_admin"},
	})
	if err != nil {
		log.Panic(err)
	}
}
