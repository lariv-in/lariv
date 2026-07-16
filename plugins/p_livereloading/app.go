package p_livereloading

import (
	"log"
	"net/url"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

// GetPlugin returns routes and metadata for [lariv.BuildAllRegistries].
// Shell head snippet registration stays in init() (see pages.go).
func GetPlugin() registry.Pair[string, lariv.Plugin] {
	u, err := url.Parse("/")
	if err != nil {
		log.Panic(err)
	}
	return registry.Pair[string, lariv.Plugin]{
		Key: "p_livereloading",
		Value: lariv.Plugin{
			Type:        lariv.PluginTypeAddon,
			Icon:        "arrow-path",
			URL:         u,
			VerboseName: "Live reload",
			Routes:      pluginStages(pluginRoutes),
		},
	}
}
