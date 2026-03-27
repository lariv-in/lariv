package p_users

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const (
	AppUrl  = "/users/"
	RoleUrl = "/roles/"
)

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_users", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "users",
		URL:         u,
		VerboseName: "Users",
	})
	if err != nil {
		log.Panic(err)
	}
}
