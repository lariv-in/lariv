package p_seer_intel

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerTablePages()
	registerDetailPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("seer_intel.IntelMenu", &components.SidebarMenu{
		Title: getters.Static("Intel"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Intel"),
				Url:   lago.RoutePath("seer_intel.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("seer_intel.IntelDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Intel: %s", getters.Any(getters.Key[string]("intel.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("All Intel"),
			Url:   lago.RoutePath("seer_intel.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("seer_intel.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("intel.ID")),
				}),
			},
		},
	})
}
