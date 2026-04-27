package p_announcements

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
	lago.RegistryPage.Register("announcements.AnnouncementMenu", &components.SidebarMenu{
		Title: getters.Static("Announcements"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All"), Url: lago.RoutePath("announcements.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create"), Url: lago.RoutePath("announcements.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("announcements.AnnouncementDetailMenu", &components.SidebarMenu{
		Title: getters.Format("%s", getters.Any(getters.Key[string]("announcement.Title"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back"), Url: lago.RoutePath("announcements.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("announcement.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("announcements.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("announcement.ID"))})},
		},
	})
}
