package p_teachers

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
	lago.RegistryPage.Register("teachers.TeacherMenu", &components.SidebarMenu{
		Title: getters.Static("Teachers"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All Teachers"), Url: lago.RoutePath("teachers.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create Teacher"), Url: lago.RoutePath("teachers.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("teachers.TeacherDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Teacher: %s", getters.Any(getters.Key[string]("teacher.Name"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to Teachers"), Url: lago.RoutePath("teachers.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("teacher.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("teachers.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("teacher.ID"))})},
		},
	})
}
