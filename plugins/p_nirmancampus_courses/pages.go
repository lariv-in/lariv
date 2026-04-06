package p_nirmancampus_courses

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

func registerMenuPages() {
	lago.RegistryPage.Register("courses.CourseDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Course: %s", getters.Any(getters.Key[string]("course.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to courses"),
			Url:   lago.RoutePath("courses.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Course Detail"),
				Url:   lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit Course"),
				Url:   lago.RoutePath("courses.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))}),
			},
		},
	})
}
