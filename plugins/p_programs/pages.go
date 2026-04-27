package p_programs

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
	lago.RegistryPage.Register("programs.ProgramMenu", &components.SidebarMenu{
		Title: getters.Static("Programs"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All Programs"), Url: lago.RoutePath("programs.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create Program"), Url: lago.RoutePath("programs.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("programs.ProgramDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Program: %s", getters.Any(getters.Key[string]("program.Name"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to Programs"), Url: lago.RoutePath("programs.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("program.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("programs.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("program.ID"))})},
		},
	})
}
