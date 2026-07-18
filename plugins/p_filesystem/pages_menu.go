package p_filesystem

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func pageEntriesMenus() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		{Key: "filesystem.MainMenu", Value: &components.SidebarMenu{
			Title: getters.Static("Filesystem"),
			Back: &components.SidebarMenuItem{
				Title: getters.Static("Back to All Apps"),
				Url:   lariv.RoutePath("dashboard.AppsPage", nil),
			},
			Children: []components.PageInterface{
				&components.SidebarMenuItem{Title: getters.Static("All Files"), Url: lariv.RoutePath("filesystem.ListRoute", nil), Icon: "folder-open"},
				&components.SidebarMenuItem{Title: getters.Static("Create Item"), Url: lariv.RoutePath("filesystem.CreateRoute", nil), Icon: "plus"},
				&components.SidebarMenuItem{Title: getters.Static("Bulk Upload"), Url: lariv.RoutePath("filesystem.MultiUploadRoute", nil), Icon: "arrow-up-tray"},
				&components.SidebarMenuItem{Title: getters.Static("Upload Zip"), Url: lariv.RoutePath("filesystem.ZipUploadRoute", nil), Icon: "archive-box"},
			},
		}},
		{Key: "filesystem.VNodeMenu", Value: &components.SidebarMenu{
			Title: currentVNodeTitle(),
			Back: &components.SidebarMenuItem{
				Title: getters.Static("Back"),
				Url:   currentVNodeBackRoute(),
			},
			Children: []components.PageInterface{
				&components.SidebarMenuItem{Title: getters.Static("View Details"), Url: lariv.RoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("vnode.ID")),
				}), Icon: "eye"},
				&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lariv.RoutePath("filesystem.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("vnode.ID")),
				}), Icon: "pencil-square"},
				&components.SidebarMenuItem{Title: getters.Static("Move"), Url: lariv.RoutePath("filesystem.MoveRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("vnode.ID")),
				}), Icon: "arrow-right-circle"},
				&components.ShowIf{
					Getter: getters.Any(getters.Key[bool]("vnode.IsDirectory")),
					Children: []components.PageInterface{
						&components.SidebarMenuItem{Title: getters.Static("Browse Contents"), Url: lariv.RoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
							"parent_id": getters.Any(getters.Key[uint]("vnode.ID")),
						}), Icon: "folder-open"},
						&components.SidebarMenuItem{Title: getters.Static("Add New Item"), Url: lariv.RoutePath("filesystem.CreateChildRoute", map[string]getters.Getter[any]{
							"parent_id": getters.Any(getters.Key[uint]("vnode.ID")),
						}), Icon: "plus"},
						&components.SidebarMenuItem{Title: getters.Static("Bulk Upload"), Url: lariv.RoutePath("filesystem.MultiUploadChildRoute", map[string]getters.Getter[any]{
							"parent_id": getters.Any(getters.Key[uint]("vnode.ID")),
						}), Icon: "arrow-up-tray"},
						&components.SidebarMenuItem{Title: getters.Static("Upload Zip"), Url: lariv.RoutePath("filesystem.ZipUploadChildRoute", map[string]getters.Getter[any]{
							"parent_id": getters.Any(getters.Key[uint]("vnode.ID")),
						}), Icon: "archive-box"},
					},
				},
			},
		}},
	}
}
