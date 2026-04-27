package p_students

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
			&components.SidebarMenuItem{
				Title: getters.Static("Create Student"),
				Url:   lago.RoutePath("students.CreateRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("students.StudentDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Student: %s", getters.Any(getters.Key[string]("student.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Students"),
			Url:   lago.RoutePath("students.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("student.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("students.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("student.ID")),
				}),
			},
		},
	})
}
