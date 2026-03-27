package p_filesystem

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppUrl = "/filesystem/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	err = lago.RegistryPlugin.Register("p_filesystem", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "folder",
		URL:         u,
		VerboseName: "Filesystem",
		Roles:       []string{"superuser", "admin"},
	})
	if err != nil {
		log.Panic(err)
	}
}
