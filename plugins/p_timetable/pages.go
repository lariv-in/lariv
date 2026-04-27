package p_timetable

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
	lago.RegistryPage.Register("timetable.TimetableSlotMenu", &components.SidebarMenu{
		Title: getters.Static("Timetable"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All slots"), Url: lago.RoutePath("timetable.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create"), Url: lago.RoutePath("timetable.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("timetable.TimetableSlotDetailMenu", &components.SidebarMenu{
		Title: getters.Format("%s", getters.Any(getters.Key[string]("timetable_slot.Label"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back"), Url: lago.RoutePath("timetable.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("timetable.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("timetable_slot.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("timetable.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("timetable_slot.ID"))})},
		},
	})
}
