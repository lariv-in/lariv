package p_totschool_users

import (
	"github.com/lariv-in/lago"
)

func init() {
	lago.RegistryPlugin.Patch("p_users", func(plugin lago.Plugin) lago.Plugin {
		plugin.Roles = []string{"superuser", "totschool_admin"}
		return plugin
	})

	lago.RegistryPlugin.Patch("p_otp", func(plugin lago.Plugin) lago.Plugin {
		plugin.Roles = []string{"superuser", "totschool_admin"}
		return plugin
	})
}
