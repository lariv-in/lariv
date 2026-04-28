package p_finances

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
	lago.RegistryPage.Register("finances.StudentChargeMenu", &components.SidebarMenu{
		Title: getters.Static("Finances"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All charges"), Url: lago.RoutePath("finances.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create charge"), Url: lago.RoutePath("finances.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("finances.StudentChargeDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Charge #%d", getters.Any(getters.Key[uint]("student_charge.ID"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back"), Url: lago.RoutePath("finances.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("finances.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student_charge.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("finances.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student_charge.ID"))})},
		},
	})
}
