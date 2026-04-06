package p_filesystem

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerMenus() {
	lago.RegistryPage.Register("filesystem.MainMenu", &components.SidebarMenu{
		Title: getters.Static("Filesystem"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All Files"), Url: lago.RoutePath("filesystem.ListRoute", nil), Icon: "folder-open"},
			&components.SidebarMenuItem{Title: getters.Static("Create Item"), Url: lago.RoutePath("filesystem.CreateRoute", nil), Icon: "plus"},
			&components.SidebarMenuItem{Title: getters.Static("Bulk Upload"), Url: lago.RoutePath("filesystem.MultiUploadRoute", nil), Icon: "arrow-up-tray"},
		},
	})

	lago.RegistryPage.Register("filesystem.VNodeMenu", &components.SidebarMenu{
		Title: currentVNodeTitle(),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back"),
			Url:   currentVNodeBackRoute(),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("View Details"), Url: currentVNodeDetailRoute(), Icon: "eye"},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: currentVNodeEditRoute(), Icon: "pencil-square"},
			&components.SidebarMenuItem{Title: getters.Static("Move"), Url: currentVNodeMoveRoute(), Icon: "arrow-right-circle"},
			&components.ShowIf{
				Getter: currentVNodeIsDirectory(),
				Children: []components.PageInterface{
					&components.SidebarMenuItem{Title: getters.Static("Browse Contents"), Url: currentVNodeBrowseRoute(), Icon: "folder-open"},
					&components.SidebarMenuItem{Title: getters.Static("Add New Item"), Url: currentVNodeCreateChildRoute(), Icon: "plus"},
					&components.SidebarMenuItem{Title: getters.Static("Bulk Upload"), Url: currentVNodeUploadChildRoute(), Icon: "arrow-up-tray"},
				},
			},
		},
	})
}
