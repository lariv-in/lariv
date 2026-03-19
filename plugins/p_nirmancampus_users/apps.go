package p_nirmancampus_users

import (
	"github.com/lariv-in/lago"
)

func init() {
	lago.RegistryPlugin.Patch("p_users", func(plugin lago.Plugin) lago.Plugin {
		plugin.Roles = []string{"superuser", "nirmancampus_admin"}
		return plugin
	})
}
