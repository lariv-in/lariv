package p_totschool_appointments

import (
	"context"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func registerFilter() {
	lago.RegistryPage.Register("appointments.AppointmentFilter", components.FormComponent[Appointment]{
		Attr: getters.FormBoostedGet(lago.RoutePath("appointments.ListRoute", nil)),

		ChildrenInput: []components.PageInterface{
			components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
				},
			},
			components.ContainerError{
				Error: getters.Key[error]("$error.Location"),
				Children: []components.PageInterface{
					components.InputText{Label: "Location", Name: "Location", Getter: getters.Key[string]("$get.Location")},
				},
			},
			components.ContainerError{
				Error: getters.Key[error]("$error.Date"),
				Children: []components.PageInterface{
					components.InputDate{Label: "Date", Name: "Date", Getter: getters.Key[time.Time]("$get.Date")},
				},
			},
			components.InputCheckbox{Label: "Overlaps Only", Name: "Overlapping", Getter: getters.Key[bool]("$get.Overlapping")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply Filters"},
				components.ButtonClear{Label: "Clear"},
			}},
		},
	})
}

func registerTable() {
	lago.RegistryPage.Register("appointments.AppointmentTable", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentMenu"}},
		Children: []components.PageInterface{
			components.DataTable[Appointment]{
				UID:      "appointment-table",
				Data:     getters.Key[components.ObjectList[Appointment]]("appointments"),
				Title:    "Appointments",
				Subtitle: "List of appointments",
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "appointments.AppointmentFilter"}},
					&components.TableButtonCreate{Link: lago.RoutePath("appointments.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
					{Label: "Location", Name: "Location", Children: []components.PageInterface{components.FieldText{Getter: getters.Key[string]("$row.Location")}}},
					{Label: "Phone", Name: "Phone", Children: []components.PageInterface{components.FieldPhone{Getter: getters.Key[string]("$row.Phone")}}},
					{Label: "Date & Time", Name: "Datetime", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Datetime")}}},
					{Label: "Created By", Name: "CreatedBy", Children: []components.PageInterface{components.FieldText{Getter: getters.ForeignKey[p_users.User, uint, string](getters.Key[uint]("$row.CreatedByID"), "Name")}}},
					{Label: "Created At", Name: "CreatedAt", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.Key[time.Time]("$row.CreatedAt")}}},
				},
			},
		},
	})
}

func registerSelectionPages() {
	lago.RegistryPage.Register("appointments.AppointmentSelectionTable", components.Modal{
		UID: "appointment-selection-modal",
		Children: []components.PageInterface{
			components.DataTable[Appointment]{
				UID:     "appointment-selection-table",
				Title:   "Select Appointment",
				Data:    getters.Key[components.ObjectList[Appointment]]("appointments"),
				RowAttr: getters.RowAttrSelect("appointment", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "appointments.AppointmentFilter"}},
				},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
					{Label: "Location", Name: "Location", Children: []components.PageInterface{components.FieldText{Getter: getters.Key[string]("$row.Location")}}},
					{Label: "Phone", Name: "Phone", Children: []components.PageInterface{components.FieldText{Getter: getters.Key[string]("$row.Phone")}}},
					{Label: "Date & Time", Name: "Datetime", Children: []components.PageInterface{components.FieldText{Getter: getters.Key[string]("$row.Datetime")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("appointments.AppointmentCardTimelineFilter", components.FormComponent[Appointment]{
		Attr: getters.FormBoostedGet(lago.RoutePath("appointments.CardTimelineRoute", nil)),

		ChildrenInput: []components.PageInterface{
			components.ContainerError{
				Error: getters.Key[error]("$error.Date"),
				Children: []components.PageInterface{
					components.InputDate{Label: "Date", Name: "Date", Getter: getters.IfOrElse(getters.Key[time.Time]("$get.Date"), func(ctx context.Context) (time.Time, error) {
						return time.Now(), nil
					})},
				},
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply Filters"},
				components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("appointments.AppointmentCardTimeline", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentMenu"}},
		Children: []components.PageInterface{
			components.ButtonLink{Label: "Create New Appointment", Classes: "btn mb-4", Link: lago.RoutePath("appointments.CreateRoute", nil)},
			components.Timeline[Appointment]{
				UID:             "appointment-timeline",
				Title:           "Appointments Timeline",
				Data:            getters.Key[components.ObjectList[Appointment]]("appointments"),
				FilterComponent: lago.DynamicPage{Name: "appointments.AppointmentCardTimelineFilter"},
				OnClick:         lago.RoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))}),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Children: []components.PageInterface{
							components.FieldText{Classes: "font-bold", Getter: getters.Key[string]("$row.Name")},
							components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Datetime"), Classes: "text-sm font-medium whitespace-nowrap"},
							components.FieldText{Classes: "text-sm", Getter: getters.Key[string]("$row.Location")},
							components.FieldPhone{Classes: "text-sm", Getter: getters.Key[string]("$row.Phone")},
							components.ShowIf{Getter: getters.Any(getters.Key[string]("$row.Remarks")), Children: []components.PageInterface{
								components.FieldText{Getter: getters.Key[string]("$row.Remarks"), Classes: "text-sm italic"},
							}},
						},
					},
				},
			},
		},
	})
}
