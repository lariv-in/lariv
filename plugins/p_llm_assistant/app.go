package p_llm_assistant

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"
)

const AppUrl = "/llm-assistant/"

func GetPlugin() registry.Pair[string, lago.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	p := lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "sparkles",
		URL:         u,
		VerboseName: "Assistant",
		Pages:       lago.PluginStages(pluginPages),
		Views:       lago.PluginStages(pluginViews),
		Routes:      lago.PluginStages(pluginRoutes),
		Configs:     lago.PluginStages(pluginConfigs),
		DBInitHooks: lago.PluginStages(pluginDBInitHooks),
	}
	return registry.Pair[string, lago.Plugin]{Key: "p_llm_assistant", Value: p}
}
