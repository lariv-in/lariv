package p_totschool_appointments

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
)

const AppUrl = "/appointments/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	err = lago.RegistryPlugin.Register("p_totschool_appointments", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "calendar-days",
		URL:         u,
		VerboseName: "Appointments",
	})
	if err != nil {
		log.Panic(err)
	}
}
