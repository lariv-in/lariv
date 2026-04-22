package p_seer_aisstream

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerAISStreamMenuPages()
	registerAISStreamMapPages()
}

func registerAISStreamMenuPages() {
	lago.RegistryPage.Register("seer_aisstream.Menu", &components.SidebarMenu{
		Title: getters.Static("AISStream"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&aisStreamMapSidebarLink{Page: components.Page{Key: "seer_aisstream.MenuMapLink"}},
		},
	})
}
