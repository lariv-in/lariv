package p_forums

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
	lago.RegistryPage.Register("forums.ForumThreadMenu", &components.SidebarMenu{
		Title: getters.Static("Forums"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Threads"), Url: lago.RoutePath("forums.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("New thread"), Url: lago.RoutePath("forums.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("forums.ForumThreadDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Thread: %s", getters.Any(getters.Key[string]("forum_thread.Title"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back"), Url: lago.RoutePath("forums.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("forums.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("forum_thread.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("forums.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("forum_thread.ID"))})},
		},
	})
}
