package p_nirmancampus_students

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_students"
)

func init() {
	u, err := url.Parse(p_students.AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_students", lago.Plugin{
		Type:        lago.PluginTypeAddon,
		Icon:        "link",
		URL:         u,
		VerboseName: "Nirmancampus Students",
	})
	if err != nil {
		log.Panic(err)
	}
}

