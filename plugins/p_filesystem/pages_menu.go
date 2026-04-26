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
			&components.SidebarMenuItem{Title: getters.Static("View Details"), Url: lago.RoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("vnode.ID")),
			}), Icon: "eye"},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("filesystem.UpdateRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("vnode.ID")),
			}), Icon: "pencil-square"},
			&components.SidebarMenuItem{Title: getters.Static("Move"), Url: lago.RoutePath("filesystem.MoveRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("vnode.ID")),
			}), Icon: "arrow-right-circle"},
			&components.ShowIf{
				Getter: getters.Any(getters.Key[bool]("vnode.IsDirectory")),
				Children: []components.PageInterface{
					&components.SidebarMenuItem{Title: getters.Static("Browse Contents"), Url: lago.RoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
						"parent_id": getters.Any(getters.Key[uint]("vnode.ID")),
					}), Icon: "folder-open"},
					&components.SidebarMenuItem{Title: getters.Static("Add New Item"), Url: lago.RoutePath("filesystem.CreateChildRoute", map[string]getters.Getter[any]{
						"parent_id": getters.Any(getters.Key[uint]("vnode.ID")),
					}), Icon: "plus"},
					&components.SidebarMenuItem{Title: getters.Static("Bulk Upload"), Url: lago.RoutePath("filesystem.MultiUploadChildRoute", map[string]getters.Getter[any]{
						"parent_id": getters.Any(getters.Key[uint]("vnode.ID")),
					}), Icon: "arrow-up-tray"},
				},
			},
		},
	})
}
