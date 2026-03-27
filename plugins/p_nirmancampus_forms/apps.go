package p_nirmancampus_forms

import (
	"github.com/lariv-in/lago/lago"
)

func init() {
	// Dashboard apps grid: only admin (and superuser, who bypasses role filtering) sees the Forms app.
	lago.RegistryPlugin.Patch("forms", func(plugin lago.Plugin) lago.Plugin {
		plugin.Roles = []string{"admin"}
		return plugin
	})
}
