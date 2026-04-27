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
	registerExamPages()
}

// --- Menus ---

func registerMenuPages() {
	lago.RegistryPage.Register("sessions.SessionMenu", &components.SidebarMenu{
		Title: getters.Static("Sessions"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Sessions"),
				Url:   lago.RoutePath("sessions.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("sessions.SessionDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Admission: %s", getters.Any(getters.Key[string]("session.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all sessions"),
			Url:   lago.RoutePath("sessions.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("session.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("sessions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("session.ID")),
				}),
			},
		},
	})
}
