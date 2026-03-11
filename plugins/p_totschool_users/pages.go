package p_totschool_users

import (
	"github.com/lariv-in/lago"
)

func init() {
	lago.RegistryPlugins.Patch("p_users", func(plugin lago.Plugin) lago.Plugin {
		plugin.RenderKeys = []string{"superuser", "totschool_admin"}
		return plugin
	})

	lago.RegistryPlugins.Patch("preferences", func(plugin lago.Plugin) lago.Plugin {
		plugin.RenderKeys = []string{"superuser", "totschool_admin"}
		return plugin
	})
}
