package p_events

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
	lago.RegistryPage.Register("events.SchoolEventMenu", &components.SidebarMenu{
		Title: getters.Static("Events"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All events"), Url: lago.RoutePath("events.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create event"), Url: lago.RoutePath("events.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("events.SchoolEventDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Event: %s", getters.Any(getters.Key[string]("school_event.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to events"), Url: lago.RoutePath("events.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("events.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("school_event.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("events.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("school_event.ID"))})},
		},
	})
}
