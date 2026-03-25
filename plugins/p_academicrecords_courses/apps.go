package p_academicrecords_courses

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_academicrecords"
)

func init() {
	u, err := url.Parse(p_academicrecords.AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_academicrecords_courses", lago.Plugin{
		Type:        lago.PluginTypeAddon,
		Icon:        "link",
		URL:         u,
		VerboseName: "Academic Records Courses",
	})
	if err != nil {
		log.Panic(err)
	}
}
