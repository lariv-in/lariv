package p_nirmancampus_students

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
	registerStudentUserPickPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("students.StudentMenu", &components.SidebarMenu{
		Title: getters.Static("Students"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Students"),
				Url:   lago.RoutePath("students.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("students.StudentDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Student: %s", getters.Any(getters.Key[string]("student.User.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Students"),
			Url:   lago.RoutePath("students.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Student Detail"),
				Url: lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("student.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit Student"),
				Url: lago.RoutePath("students.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("student.ID")),
				}),
			},
		},
	})
}
