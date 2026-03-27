package p_nirmancampus_academicrecords

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/academicrecords/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_academicrecords", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "book-open",
		URL:         u,
		VerboseName: "Academic Records",
	})
	if err != nil {
		log.Panic(err)
	}
}
