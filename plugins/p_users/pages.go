package p_users

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerAuthPages()
	registerSelectionPages()
	registerRolePages()
}

// --- Menus ---

func registerMenuPages() {
	lago.RegistryPage.Register("users.UserMenu", &components.SidebarMenu{
		Title: getters.Static("Users"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Home"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Users"),
				Url:   lago.RoutePath("users.ListRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Roles"),
				Url:   lago.RoutePath("users.RoleListRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("users.UserDetailMenu", &components.SidebarMenu{
		Title: getters.Format("User: %s", getters.Any(getters.Key[string]("user.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Users"),
			Url:   lago.RoutePath("users.ListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("User Detail"),
				Url: lago.RoutePath("users.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("user.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit User"),
				Url: lago.RoutePath("users.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("user.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Change Password"),
				Url: lago.RoutePath("users.ChangePasswordRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("user.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("users.UserSelfMenu", &components.SidebarMenu{
		Title: getters.Format("My account: %s", getters.Any(getters.Key[string]("user.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Home"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("My Profile"),
				Url:   lago.RoutePath("users.SelfDetailRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit My Profile"),
				Url:   lago.RoutePath("users.SelfUpdateRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Change Password"),
				Url:   lago.RoutePath("users.SelfChangePasswordRoute", nil),
			},
		},
	})
}
