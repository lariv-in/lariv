package p_users

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func pageEntriesMenus() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		{Key: "p_users.UserMenu", Value: &components.SidebarMenu{
			Title: getters.Static("Users"),
			Back: &components.SidebarMenuItem{
				Title: getters.Static("Back to Home"),
				Url:   lariv.RoutePath("dashboard.AppsPage", nil),
			},
			Children: []components.PageInterface{
				&components.SidebarMenuItem{
					Title: getters.Static("All Users"),
					Url:   lariv.RoutePath("p_users.ListRoute", nil),
				},
				&components.SidebarMenuItem{
					Title: getters.Static("Roles"),
					Url:   lariv.RoutePath("p_users.RoleListRoute", nil),
				},
			},
		}},
		{Key: "p_users.UserDetailMenu", Value: &components.SidebarMenu{
			Title: getters.Format("User: %s", getters.Any(getters.Key[string]("user.Name"))),
			Back: &components.SidebarMenuItem{
				Title: getters.Static("Back to All Users"),
				Url:   lariv.RoutePath("p_users.ListRoute", nil),
			},
			Children: []components.PageInterface{
				&components.SidebarMenuItem{
					Title: getters.Static("User Detail"),
					Url: lariv.RoutePath("p_users.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("user.ID")),
					}),
				},
				&components.SidebarMenuItem{
					Title: getters.Static("Edit User"),
					Url: lariv.RoutePath("p_users.UpdateRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("user.ID")),
					}),
				},
				&components.SidebarMenuItem{
					Title: getters.Static("Change Password"),
					Url: lariv.RoutePath("p_users.ChangePasswordRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("user.ID")),
					}),
				},
			},
		}},
		{Key: "p_users.UserSelfMenu", Value: &components.SidebarMenu{
			Title: getters.Format("My account: %s", getters.Any(getters.Key[string]("user.Name"))),
			Back: &components.SidebarMenuItem{
				Title: getters.Static("Back to Home"),
				Url:   lariv.RoutePath("dashboard.AppsPage", nil),
			},
			Children: []components.PageInterface{
				&components.SidebarMenuItem{
					Title: getters.Static("My Profile"),
					Url:   lariv.RoutePath("p_users.SelfDetailRoute", nil),
				},
				&components.SidebarMenuItem{
					Title: getters.Static("Edit My Profile"),
					Url:   lariv.RoutePath("p_users.SelfUpdateRoute", nil),
				},
				&components.SidebarMenuItem{
					Title: getters.Static("Change Password"),
					Url:   lariv.RoutePath("p_users.SelfChangePasswordRoute", nil),
				},
			},
		}},
	}
}
