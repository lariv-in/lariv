package p_sessions

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("sessions.ClassSessionMenu", &components.SidebarMenu{
		Title: getters.Static("Sessions"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All sessions"), Url: lago.RoutePath("sessions.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create"), Url: lago.RoutePath("sessions.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("sessions.ClassSessionDetailMenu", &components.SidebarMenu{
		Title: getters.Format("%s", getters.Any(getters.Key[string]("class_session.Title"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back"), Url: lago.RoutePath("sessions.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("class_session.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("sessions.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("class_session.ID"))})},
		},
	})
}
