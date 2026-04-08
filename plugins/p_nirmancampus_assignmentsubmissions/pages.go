package p_nirmancampus_assignmentsubmissions

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
}

func registerMenuPages() {
	lago.RegistryPage.Register("assignmentsubmissions.ListMenu", &components.SidebarMenu{
		Title: getters.Static("Assignment Submissions"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All submissions"),
				Url:   lago.RoutePath("assignmentsubmissions.DefaultRoute", nil),
			},
		},
	})

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
