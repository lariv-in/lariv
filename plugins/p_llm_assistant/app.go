package p_llm_assistant

import (
	"log"
	"net/url"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

const AppUrl = "/llm-assistant/"

func GetPlugin() registry.Pair[string, lariv.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	p := lariv.Plugin{
		Type:        lariv.PluginTypeApp,
		Icon:        "sparkles",
		URL:         u,
		VerboseName: "Assistant",
		Pages:       lariv.PluginStages(pluginPages),
		Views:       lariv.PluginStages(pluginViews),
		Routes:      lariv.PluginStages(pluginRoutes),
		Configs:     lariv.PluginStages(pluginConfigs),
		DBInitHooks: lariv.PluginStages(pluginDBInitHooks),
	}
	return registry.Pair[string, lariv.Plugin]{Key: "p_llm_assistant", Value: p}
}
