package p_totschool_appointments

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenus()
	registerFilter()
	registerForms()
	registerTable()
	registerDetail()
	registerModal()
	registerDelete()
	registerSelectionPages()
}

func registerMenus() {
	lago.RegistryPage.Register("appointments.AppointmentMenu", components.SidebarMenu{
		Title: getters.Static("Appointments"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{Title: getters.Static("All Appointments"), Url: lago.RoutePath("appointments.ListRoute", nil)},
			components.SidebarMenuItem{Title: getters.Static("Appointments Timeline"), Url: lago.RoutePath("appointments.CardTimelineRoute", nil)},
			components.SidebarMenuItem{Title: getters.Static("Create Appointment"), Url: lago.RoutePath("appointments.CreateRoute", nil)},
		},
	})

	lago.RegistryPage.Register("appointments.AppointmentDetailMenu", components.SidebarMenu{
		Title: getters.Format("Appointment: %s", getters.Any(getters.Key[string]("appointment.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all Appointments"),
			Url:   lago.RoutePath("appointments.ListRoute", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{Title: getters.Static("Appointment Detail"), Url: lago.RoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("appointment.ID"))})},
			components.SidebarMenuItem{Title: getters.Static("Edit Appointment"), Url: lago.RoutePath("appointments.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("appointment.ID"))})},
		},
	})
}
