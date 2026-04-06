package p_nirmancampus_announcements

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
}

// --- Menus ---

func registerMenuPages() {
	lago.RegistryPage.Register("announcements.AnnouncementMenu", &components.SidebarMenu{
		Title: getters.Static("Announcements"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Announcements"),
				Url:   lago.RoutePath("announcements.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("announcements.AnnouncementDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Announcement: %s", getters.Any(getters.Key[string]("announcement.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all Announcements"),
			Url:   lago.RoutePath("announcements.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Announcement Detail"),
				Url: lago.RoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("announcement.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit Announcement"),
				Url: lago.RoutePath("announcements.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("announcement.ID")),
				}),
			},
		},
	})
}
