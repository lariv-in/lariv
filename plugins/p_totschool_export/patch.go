package p_totschool_export

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

// exportAppRoles: dashboard tile + HTTP views ([RoleAuthorizationLayer] still
// allows IsSuperuser). AppsGrid skips role filter when $role is superuser.
var exportAppRoles = []string{"admin"}

// exportMenuRoles: SidebarMenu uses components.Render ($role string); superuser
// must appear here or sidebar stays empty while export routes still work.
var exportMenuRoles = []string{"admin", "superuser"}

var exportRoleLayer = p_users.RoleAuthorizationLayer{Roles: exportAppRoles}

func init() {
	lago.RegistryPlugin.Patch("p_export", func(plugin lago.Plugin) lago.Plugin {
		plugin.Roles = append([]string(nil), exportAppRoles...)
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

	lago.RegistryPage.Patch("export.Page", func(page components.PageInterface) components.PageInterface {
		shell, ok := page.(*components.ShellScaffold)
		if !ok {
			return page
		}
		shell.Roles = append([]string(nil), exportMenuRoles...)
		return shell
	})

	lago.RegistryView.Patch("export.PageView", func(v *views.View) *views.View {
		return v.InsertLayerAfter("users.auth", "totschool_export.role", exportRoleLayer)
	})

	lago.RegistryView.Patch("export.DownloadView", func(v *views.View) *views.View {
		return v.InsertLayerAfter("users.auth", "totschool_export.role", exportRoleLayer)
	})
}
