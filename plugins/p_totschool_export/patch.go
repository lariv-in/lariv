package p_totschool_export

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

// exportDashboardRoles limits the Export app tile on the dashboard to these role
// names ($role). Superuser is handled separately in AppsGrid (always shown).
var exportDashboardRoles = []string{"admin", "totschool_admin"}

// exportMenuRoles gate SidebarMenu via components.Render ($role match).
var exportMenuRoles = []string{"superuser", "admin", "totschool_admin"}

var exportRoleLayer = p_users.RoleAuthorizationLayer{Roles: exportDashboardRoles}

func init() {
	lago.RegistryPlugin.Patch("p_export", func(plugin lago.Plugin) lago.Plugin {
		plugin.Roles = append([]string(nil), exportDashboardRoles...)
		return plugin
	})

	lago.RegistryPage.Patch("export.Menu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Roles = append([]string(nil), exportMenuRoles...)
		return menu
	})

	lago.RegistryView.Patch("export.PageView", func(v *views.View) *views.View {
		return v.InsertLayerAfter("users.auth", "totschool_export.role", exportRoleLayer)
	})

	lago.RegistryView.Patch("export.DownloadView", func(v *views.View) *views.View {
		return v.InsertLayerAfter("users.auth", "totschool_export.role", exportRoleLayer)
	})
}
