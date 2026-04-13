package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	lago.RegistryPage.Register("nirmancampus_website.AppLandingPage", &websiteAppLandingPage{})

	lago.RegistryPage.Register("nirmancampus_website.WebsiteAdminMenu", &components.SidebarMenu{
		Title: getters.Static("Website"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Home"),
				Url:   lago.RoutePath("nirmancampus_website.AppLandingRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Student Zone Sections"),
				Url:   lago.RoutePath("nirmancampus_website.StudentZoneAdminDefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Student Zone Items"),
				Url:   lago.RoutePath("nirmancampus_website.StudentZoneAdminItemListRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Important Links"),
				Url:   lago.RoutePath("nirmancampus_website.ImportantLinksDefaultRoute", nil),
			},
		},
	})
}
