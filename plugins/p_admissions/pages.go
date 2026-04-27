package p_admissions

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
	lago.RegistryPage.Register("admissions.ApplicationMenu", &components.SidebarMenu{
		Title: getters.Static("Admissions"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All Applications"), Url: lago.RoutePath("admissions.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create Application"), Url: lago.RoutePath("admissions.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("admissions.ApplicationDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Application: %s", getters.Any(getters.Key[string]("application.ApplicantName"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to Applications"), Url: lago.RoutePath("admissions.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("admissions.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("application.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("admissions.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("application.ID"))})},
		},
	})
}
