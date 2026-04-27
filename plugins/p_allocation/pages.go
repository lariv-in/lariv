package p_allocation

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
	lago.RegistryPage.Register("allocation.CourseTeacherAssignmentMenu", &components.SidebarMenu{
		Title: getters.Static("Allocation"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All rows"), Url: lago.RoutePath("allocation.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create"), Url: lago.RoutePath("allocation.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("allocation.CourseTeacherAssignmentDetailMenu", &components.SidebarMenu{
		Title: getters.Static("Teacher–course"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back"), Url: lago.RoutePath("allocation.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("allocation.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("allocation.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("allocation.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("allocation.ID"))})},
		},
	})
}
