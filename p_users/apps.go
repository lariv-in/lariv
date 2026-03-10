package p_users

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
)

const AppUrl = "/users/"
const RoleUrl = "/roles/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugins.Register("p_users", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "users",
		Url:         u,
		VerboseName: "Users",
	})
	if err != nil {
		log.Panic(err)
	}
}
