package p_totschool_users

import (
	"context"
	"sort"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	pcomps "github.com/lariv-in/p_dashboard/components"
	"github.com/lariv-in/p_users"
)

func filteredAppsGetter(ctx context.Context) any {
	user, ok := ctx.Value("$user").(p_users.User)

	allPlugins := lago.RegistryPlugins.All()
	var apps []lago.Plugin
	for name, plugin := range *allPlugins {
		if plugin.Type != lago.PluginTypeApp {
			continue
		}
		if ok && !user.IsSuperuser && user.Role.Name != "totschool_admin" {
			if name == "p_users" || name == "preferences" {
				continue
			}
		}
		apps = append(apps, plugin)
	}
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].VerboseName < apps[j].VerboseName
	})
	return apps
}

func init() {
	lago.RegistryPage.Patch("dashboard.AppsPage", func(oldPage components.PageInterface) components.PageInterface {
		scaffold, ok := oldPage.(components.ShellTopbarScaffold)
		if !ok {
			return oldPage
		}
		if len(scaffold.Children) == 0 {
			return oldPage
		}
		layout, ok := scaffold.Children[0].(components.LayoutSimple)
		if !ok {
			return oldPage
		}

		// Replace AppsGrid with one that filters apps
		layout.Children = []components.PageInterface{
			components.ContainerRow{
				Classes: "flex items-end gap-2 text-3xl font-bold text-base-content mb-2 mt-4 max-w-5xl mx-auto px-6",
				Children: []components.PageInterface{
					components.FieldText{Getter: getters.GetterStatic("Hello, ")},
					components.FieldText{Getter: getters.GetterKey("$user.Name")},
				},
			},
			pcomps.AppsGrid{Apps: filteredAppsGetter},
		}

		scaffold.Children[0] = layout
		return scaffold
	})
}
