package p_seer_runners

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerTablePages()
	registerDetailPages()
	registerFormPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("seer_runners.RunnerMenu", &components.SidebarMenu{
		Title: getters.Static("Runners"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Runners"),
				Url:   lago.RoutePath("seer_runners.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Create runner"),
				Url:   lago.RoutePath("seer_runners.CreateRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("seer_runners.RunnerDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Runner #%d", getters.Any(getters.Key[uint]("runner.ID"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("All Runners"),
			Url:   lago.RoutePath("seer_runners.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("seer_runners.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("runner.ID")),
				}),
			},
		},
	})
}
