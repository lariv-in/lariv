package p_assessments

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/assessments/"

// ExamsURL is CRUD for Assessment (exam definitions); AppURL stays grade-entry list.
const ExamsURL = AppURL + "exams/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_assessments", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "chart-bar",
		URL:         u,
		VerboseName: "Assessments",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
