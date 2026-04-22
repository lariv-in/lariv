package p_seer_websites

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerWebsiteMenuPages()
	registerWebsiteSourceMenus()
	registerWebsiteSourcePages()
	registerWebsiteSourceFormPages()
	registerWebsiteRunnerPages()
	registerWebsiteRunnerWorkerPoolViews()

	registerWebsiteTablePages()
	registerWebsiteFormPages()
	registerWebsiteDetailPages()
}

func registerWebsiteSourceMenus() {
	lago.RegistryPage.Register("seer_websites.WebsiteSourceDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Website source #%d", getters.Any(getters.Key[uint]("websiteSource.ID"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Sources"),
			Url:   lago.RoutePath("seer_websites.WebsiteSourceListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("seer_websites.WebsiteSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("websiteSource.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("seer_websites.WebsiteSourceUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("websiteSource.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("seer_websites.WebsiteRunnerDetailMenu", &components.SidebarMenu{
		Title: getters.Key[string]("websiteRunner.Name"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Workers"),
			Url:   lago.RoutePath("seer_websites.WebsiteRunnerListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("seer_websites.WebsiteRunnerDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("websiteRunner.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("seer_websites.WebsiteRunnerUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("websiteRunner.ID")),
				}),
			},
		},
	})
}

func registerWebsiteMenuPages() {
	lago.RegistryPage.Register("seer_websites.WebsiteMenu", &components.SidebarMenu{
		Title: getters.Static("Websites"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Saved pages"),
				Url:   lago.RoutePath("seer_websites.WebsiteListRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Sources"),
				Url:   lago.RoutePath("seer_websites.WebsiteSourceListRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Workers"),
				Url:   lago.RoutePath("seer_websites.WebsiteRunnerListRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Add from URL"),
				Url:   lago.RoutePath("seer_websites.WebsiteAddRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("seer_websites.WebsiteDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Website #%d", getters.Any(getters.Key[uint]("website.ID"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Saved pages"),
			Url:   lago.RoutePath("seer_websites.WebsiteListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("seer_websites.WebsiteDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("website.ID")),
				}),
			},
		},
	})
}
