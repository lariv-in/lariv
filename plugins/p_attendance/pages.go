package p_attendance

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
	lago.RegistryPage.Register("attendance.AttendanceMarkMenu", &components.SidebarMenu{
		Title: getters.Static("Attendance"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All marks"), Url: lago.RoutePath("attendance.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create"), Url: lago.RoutePath("attendance.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("attendance.AttendanceMarkDetailMenu", &components.SidebarMenu{
		Title: getters.Static("Attendance mark"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back"), Url: lago.RoutePath("attendance.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("attendance.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("attendance_mark.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("attendance.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("attendance_mark.ID"))})},
		},
	})
}
