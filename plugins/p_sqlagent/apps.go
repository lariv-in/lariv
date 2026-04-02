package sqlagent

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_sqlagent", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "chat-bubble-left-right",
		URL:         u,
		VerboseName: "SQL Agent",
	}); err != nil {
		log.Panic(err)
	}
}
