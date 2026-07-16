package p_export

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func pluginPages() lariv.PluginFeatures[components.PageInterface] {
	return lariv.PluginFeatures[components.PageInterface]{
		Entries: []registry.Pair[string, components.PageInterface]{
			{Key: "export.Menu", Value: components.SidebarMenu{
				Title: getters.Static("Export"),
				Back: &components.SidebarMenuItem{
					Title: getters.Static("Back to All Apps"),
					Url:   lariv.RoutePath("dashboard.AppsPage", nil),
				},
				Children: []components.PageInterface{
					components.SidebarMenuItem{
						Title:  getters.Static("XLSX Export"),
						Url:    lariv.RoutePath("export.PageRoute", nil),
						Active: true,
					},
				},
			}},
			{Key: "export.Page", Value: &components.ShellScaffold{
				Sidebar: []components.PageInterface{
					lariv.DynamicPage{Name: "export.Menu"},
				},
				Children: []components.PageInterface{
					exportPickerPage{},
				},
			}},
		},
	}
}
