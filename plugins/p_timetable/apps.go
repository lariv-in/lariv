package p_timetable

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/timetable/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_timetable", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "table-cells",
		URL:         u,
		VerboseName: "Timetable",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
