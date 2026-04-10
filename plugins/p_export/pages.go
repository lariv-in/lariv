package p_export

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenu()
	registerPages()
}

func registerMenu() {
	lago.RegistryPage.Register("export.Menu", components.SidebarMenu{
		Title: getters.Static("Export"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title:  getters.Static("XLSX Export"),
				Url:    lago.RoutePath("export.PageRoute", nil),
				Active: true,
			},
		},
	})
}

func registerPages() {
	lago.RegistryPage.Register("export.Page", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "export.Menu"},
		},
		Children: []components.PageInterface{
			exportPickerPage{},
		},
	})
}
