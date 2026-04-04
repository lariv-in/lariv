package p_nirmancampus_academicrecords

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
)

// AppUrl is the academic records area inside the Students app (not a standalone dashboard app).
var AppUrl = p_nirmancampus_students.AppUrl + "academicrecords/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_academicrecords", lago.Plugin{
		Type:        lago.PluginTypeAddon,
		Icon:        "book-open",
		URL:         u,
		VerboseName: "Academic Records",
		Roles:       []string{"superuser", "admin", "student"},
	})
	if err != nil {
		log.Panic(err)
	}
}
