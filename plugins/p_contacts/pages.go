package p_contacts

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

func registerMenuPages() {
	lago.RegistryPage.Register("contacts.ContactMenu", &components.SidebarMenu{
		Title: getters.Static("Contacts"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Contacts"),
				Url:   lago.RoutePath("contacts.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("contacts.ContactDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Contact: %s", getters.Any(getters.Key[string]("contact.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all Contacts"),
			Url:   lago.RoutePath("contacts.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Contact Detail"),
				Url: lago.RoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("contact.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit Contact"),
				Url: lago.RoutePath("contacts.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("contact.ID")),
				}),
			},
		},
	})
}
