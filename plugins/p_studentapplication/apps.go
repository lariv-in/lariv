package p_studentapplication

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
)

const AppUrl = "/student-applications/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_studentapplication", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "document-text",
		URL:         u,
		VerboseName: "Student applications",
	})
	if err != nil {
		log.Panic(err)
	}
}
