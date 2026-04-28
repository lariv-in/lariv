package p_syllabus

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/syllabus/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_syllabus", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "list-bullet",
		URL:         u,
		VerboseName: "Syllabus",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
