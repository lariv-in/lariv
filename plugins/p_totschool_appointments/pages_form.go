package p_totschool_appointments

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func appointmentFormFields() []components.PageInterface {
	return []components.PageInterface{
		components.ContainerError{
			Error: getters.Key[error]("$error.Name"),
			Children: []components.PageInterface{
				components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.Key[string]("$in.Name")},
			},
		},
		components.ContainerError{
			Error: getters.Key[error]("$error.Location"),
			Children: []components.PageInterface{
				components.InputTextarea{Label: "Location", Name: "Location", Required: true, Getter: getters.Key[string]("$in.Location"), Rows: 2},
			},
		},
		components.ContainerRow{Classes: "grid grid-cols-1 gap-1 md:grid-cols-2", Children: []components.PageInterface{
			components.ContainerError{
				Error: getters.Key[error]("$error.Phone"),
				Children: []components.PageInterface{
					components.InputPhone{Label: "Phone", Name: "Phone", Required: true, Getter: getters.Key[string]("$in.Phone")},
				},
			},
			components.ContainerError{
				Error: getters.Key[error]("$error.Datetime"),
				Children: []components.PageInterface{
					components.InputDatetime{Label: "Date & Time", Name: "Datetime", Required: true, Getter: getters.Key[time.Time]("$in.Datetime")},
				},
			},
		}},
		components.ContainerError{
			Error: getters.Key[error]("$error.Remarks"),
			Children: []components.PageInterface{
				components.InputTextarea{Label: "Remarks", Name: "Remarks", Getter: getters.Key[string]("$in.Remarks"), Rows: 2},
			},
		},
		components.ContainerError{
			Error: getters.Key[error]("$error.ExtraInfo"),
			Children: []components.PageInterface{
				components.InputTextarea{Label: "Extra Info (For AI)", Name: "ExtraInfo", Getter: getters.Key[string]("$in.ExtraInfo"), Rows: 2},
			},
		},
	}
}

func registerForms() {
	lago.RegistryPage.Register("appointments.AppointmentCreateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				ActionURL: lago.RoutePath("appointments.CreateRoute", nil),
				Children: []components.PageInterface{
					components.FormComponent[Appointment]{
						Attr: getters.FormBubbling(nil),

						Title:          "Create Appointment",
						Subtitle:       "Create a new appointment",
						ChildrenInput:  appointmentFormFields(),
						ChildrenAction: []components.PageInterface{components.ButtonSubmit{Label: "Save Appointment"}},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("appointments.AppointmentUpdateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				ActionURL: lago.RoutePath("appointments.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("appointment.ID"))}),
				Children: []components.PageInterface{
					components.FormComponent[Appointment]{
						Getter: getters.Key[Appointment]("appointment"),
						Attr:   getters.FormBubbling(nil),

						Title:         "Edit Appointment",
						Subtitle:      "Update appointment details",
						ChildrenInput: appointmentFormFields(),
						ChildrenAction: []components.PageInterface{
							components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									components.ButtonModalForm{
										Label:       "Delete",
										Icon:        "trash",
										Url:         lago.RoutePath("appointments.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("appointment.ID"))}),
										FormPostURL: lago.RoutePath("appointments.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("appointment.ID"))}),
										ModalUID:    "appointment-delete-modal",
										Classes:     "btn-outline btn-error btn-sm",
									},
									components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											components.ButtonSubmit{Label: "Save Appointment"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
}
