package p_nirmancampus_sessions

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
	registerSelectionPages()
}

// --- Menus ---

func registerMenuPages() {
	lago.RegistryPage.Register("sessions.SessionMenu", &components.SidebarMenu{
		Title: getters.Static("sessions"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All sessions"),
				Url:   lago.RoutePath("sessions.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("sessions.SessionDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Session: %s", getters.Any(getters.Key[string]("session.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all sessions"),
			Url:   lago.RoutePath("sessions.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Session Detail"),
				Url: lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("session.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit Session"),
				Url: lago.RoutePath("sessions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("session.ID")),
				}),
			},
		},
	})
}

// --- Filters ---
