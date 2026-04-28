package p_nirmancampus_examregistrations

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerStudentsMenuExamRegistrationsEntry()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
}

func registerStudentsMenuExamRegistrationsEntry() {
	lago.RegistryPage.Patch("students.StudentMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("All Exam Registrations"),
			Url:   lago.RoutePath("examregistrations.DefaultRoute", nil),
		})
		return menu
	})
}

func registerMenuPages() {
	lago.RegistryPage.Register("examregistrations.DetailMenu", &components.SidebarMenu{
		Title: getters.Format("Registration: %s", getters.Any(getters.Key[string]("examregistration.ExamTitle"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to exam registrations"),
			Url:   lago.RoutePath("examregistrations.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Registration detail"),
				Url: lago.RoutePath("examregistrations.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("examregistration.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit registration"),
				Url: lago.RoutePath("examregistrations.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("examregistration.ID")),
				}),
			},
		},
	})
}
