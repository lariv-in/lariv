package p_totschool_users

import (
	"github.com/lariv-in/lago/lago"
)

func init() {
	lago.RegistryPlugin.Patch("p_users", func(plugin lago.Plugin) lago.Plugin {
		plugin.Roles = []string{"superuser", "totschool_admin"}
		return plugin
	})
}
