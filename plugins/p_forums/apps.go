package p_forums

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/forums/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_forums", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "chat-bubble-left-right",
		URL:         u,
		VerboseName: "Forums",
		Roles:       []string{"superuser", "admin"},
	}); err != nil {
		log.Panic(err)
	}
}
