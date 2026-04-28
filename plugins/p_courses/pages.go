package p_courses

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
	lago.RegistryPage.Register("courses.CourseMenu", &components.SidebarMenu{
		Title: getters.Static("Courses"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All Courses"), Url: lago.RoutePath("courses.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create Course"), Url: lago.RoutePath("courses.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("courses.CourseDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Course: %s", getters.Any(getters.Key[string]("course.Name"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to Courses"), Url: lago.RoutePath("courses.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("courses.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))})},
		},
	})
}
