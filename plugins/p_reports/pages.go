package p_reports

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
	lago.RegistryPage.Register("reports.ReportDefinitionMenu", &components.SidebarMenu{
		Title: getters.Static("Reports"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Definitions"), Url: lago.RoutePath("reports.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create"), Url: lago.RoutePath("reports.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("reports.ReportDefinitionDetailMenu", &components.SidebarMenu{
		Title: getters.Format("%s", getters.Any(getters.Key[string]("report_definition.Name"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back"), Url: lago.RoutePath("reports.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("reports.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("report_definition.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("reports.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("report_definition.ID"))})},
		},
	})
}
