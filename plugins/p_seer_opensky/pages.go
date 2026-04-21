package p_seer_opensky

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerOpenSkyMenuPages()
	registerOpenSkyMapPages()
}

func registerOpenSkyMenuPages() {
	lago.RegistryPage.Register("seer_opensky.Menu", &components.SidebarMenu{
		Title: getters.Static("OpenSky"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&openskyMapSidebarLink{Page: components.Page{Key: "seer_opensky.MenuMapLink"}},
		},
	})
}
