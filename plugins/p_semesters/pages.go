package p_semesters

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
	lago.RegistryPage.Register("semesters.SemesterMenu", &components.SidebarMenu{
		Title: getters.Static("Semesters"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All Semesters"), Url: lago.RoutePath("semesters.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create Semester"), Url: lago.RoutePath("semesters.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("semesters.SemesterDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Semester: %s", getters.Any(getters.Key[string]("semester.Name"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to Semesters"), Url: lago.RoutePath("semesters.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("semesters.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("semester.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("semesters.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("semester.ID"))})},
		},
	})
}
