package p_assignments

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
}

func registerMenuPages() {
	lago.RegistryPage.Register("assignments.AssignmentMenu", &components.SidebarMenu{
		Title: getters.Static("Assignments"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All"), Url: lago.RoutePath("assignments.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create"), Url: lago.RoutePath("assignments.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("assignments.AssignmentDetailMenu", &components.SidebarMenu{
		Title: getters.Format("%s", getters.Any(getters.Key[string]("assignment.Title"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back"), Url: lago.RoutePath("assignments.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assignment.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("assignments.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assignment.ID"))})},
		},
	})
}
