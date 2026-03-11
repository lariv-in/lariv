package p_totschool_tally

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
)

const AppUrl = "/tally/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugins.Register("p_totschool_tally", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "table-cells",
		Url:         u,
		VerboseName: "Totschool Tally",
	})
	if err != nil {
		log.Panic(err)
	}
}
