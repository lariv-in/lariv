package p_nirmancampus_assignmentsubmissions

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerStudentsMenuAssignmentSubmissionsEntry()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
}

func registerStudentsMenuAssignmentSubmissionsEntry() {
	lago.RegistryPage.Patch("students.StudentMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("Assignment Submissions"),
			Url:   lago.RoutePath("assignmentsubmissions.DefaultRoute", nil),
		})
		return menu
	})
}

func registerMenuPages() {
	lago.RegistryPage.Register("assignmentsubmissions.DetailMenu", &components.SidebarMenu{
		Title: getters.Format("Submission: %s", getters.Any(getters.Key[string]("assignmentsubmission.AssignmentTitle"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to assignment submissions"),
			Url:   lago.RoutePath("assignmentsubmissions.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Submission detail"),
				Url: lago.RoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit submission"),
				Url: lago.RoutePath("assignmentsubmissions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID")),
				}),
			},
		},
	})
}
