package p_lacerate

import (
	"log"
	"log/slog"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/lacerate/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		slog.Error("lacerate: parse plugin app URL", "error", err, "url", AppUrl)
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_lacerate", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "scissors",
		URL:         u,
		VerboseName: "Lacerate",
	})
	if err != nil {
		slog.Error("lacerate: register plugin", "error", err)
		log.Panic(err)
	}
}
