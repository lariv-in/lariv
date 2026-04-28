package p_reports

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/reports/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_reports", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "document-chart-bar",
		URL:         u,
		VerboseName: "Reports",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
