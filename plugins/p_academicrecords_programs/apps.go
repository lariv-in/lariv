package p_academicrecords_programs

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_academicrecords"
)

func init() {
	u, err := url.Parse(p_academicrecords.AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_academicrecords_programs", lago.Plugin{
		Type:        lago.PluginTypeAddon,
		Icon:        "link",
		URL:         u,
		VerboseName: "Academic Records Programs",
	})
	if err != nil {
		log.Panic(err)
	}
}

