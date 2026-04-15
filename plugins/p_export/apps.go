package p_export

import (
	"log"
	"log/slog"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/export/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		slog.Error("export: parse app url", "url", AppUrl, "error", err)
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_export", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "arrow-down-tray",
		URL:         u,
		VerboseName: "Export",
	})
	if err != nil {
		slog.Error("export: register plugin", "error", err)
		log.Panic(err)
	}
}
