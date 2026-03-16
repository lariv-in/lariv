package p_totschool_appointments

import (
	"context"
	"net/http"
	"time"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
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
		Title: getters.GetterStatic("Appointments"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{Title: getters.GetterStatic("All Appointments"), Url: lago.GetterRoutePath("appointments.ListRoute", nil)},
			components.SidebarMenuItem{Title: getters.GetterStatic("Appointments Timeline"), Url: lago.GetterRoutePath("appointments.CardTimelineRoute", nil)},
			components.SidebarMenuItem{Title: getters.GetterStatic("Create Appointment"), Url: lago.GetterRoutePath("appointments.CreateRoute", nil)},
		},
	})

	lago.RegistryPage.Register("appointments.AppointmentDetailMenu", components.SidebarMenu{
		Title: getters.GetterFormat("Appointment: %s", getters.GetterAny(getters.GetterKey[string]("appointment.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Appointments"),
			Url:   lago.GetterRoutePath("appointments.ListRoute", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{Title: getters.GetterStatic("Appointment Detail"), Url: lago.GetterRoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("appointment.ID"))})},
			components.SidebarMenuItem{Title: getters.GetterStatic("Edit Appointment"), Url: lago.GetterRoutePath("appointments.UpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("appointment.ID"))})},
			components.SidebarMenuItem{Title: getters.GetterStatic("Delete Appointment"), Url: lago.GetterRoutePath("appointments.DeleteRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("appointment.ID"))})},
		},
	})
}

func registerFilter() {
	lago.RegistryPage.Register("appointments.AppointmentFilter", components.FormComponent[Appointment]{
		Url:    lago.GetterRoutePath("appointments.ListRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$get.Name")},
			components.InputText{Label: "Location", Name: "Location", Getter: getters.GetterKey[string]("$get.Location")},
			components.InputText{Label: "Date", Name: "Date", Getter: getters.GetterKey[string]("$get.Date")},
			components.InputManyToMany[[]uint]{
				Label:       "Created By",
				Name:        "CreatedBy",
				Url:         lago.GetterRoutePath("users.MultiSelectRoute", nil),
				DisplayAttr: "Name",
				Placeholder: "Select users...",
				Getter:      getters.GetterKey[[]uint]("$get.CreatedBy"),
			},
			components.InputCheckbox{Label: "Overlaps Only", Name: "Overlapping", Getter: getters.GetterKey[bool]("$get.Overlapping")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply Filters"},
				components.ButtonClear{Label: "Clear"},
			}},
		},
	})
}

func appointmentFormFields() []components.PageInterface {
	return []components.PageInterface{
		components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.GetterKey[string]("$in.Name")},
		components.InputTextarea{Label: "Location", Name: "Location", Required: true, Getter: getters.GetterKey[string]("$in.Location"), Rows: 2},
		components.ContainerRow{Classes: "grid grid-cols-1 gap-1 md:grid-cols-2", Children: []components.PageInterface{
			components.InputPhone{Label: "Phone", Name: "Phone", Required: true, Getter: getters.GetterKey[string]("$in.Phone")},
			components.InputDatetime{Label: "Date & Time", Name: "Datetime", Required: true, Getter: getters.GetterKey[time.Time]("$in.Datetime")},
		}},
		components.InputTextarea{Label: "Remarks", Name: "Remarks", Getter: getters.GetterKey[string]("$in.Remarks"), Rows: 2},
		components.InputTextarea{Label: "Extra Info (For AI)", Name: "ExtraInfo", Getter: getters.GetterKey[string]("$in.ExtraInfo"), Rows: 2},
	}
}

func registerForms() {
	lago.RegistryPage.Register("appointments.AppointmentCreateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentMenu"}},
		Children: []components.PageInterface{
			components.FormComponent[Appointment]{
				Url:            lago.GetterRoutePath("appointments.CreateRoute", nil),
				Method:         http.MethodPost,
				Title:          "Create Appointment",
				Subtitle:       "Create a new appointment",
				ChildrenInput:  appointmentFormFields(),
				ChildrenAction: []components.PageInterface{components.ButtonSubmit{Label: "Save Appointment"}},
			},
		},
	})

	lago.RegistryPage.Register("appointments.AppointmentUpdateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentDetailMenu"}},
		Children: []components.PageInterface{
			components.FormComponent[Appointment]{
				Getter:         getters.GetterKey[Appointment]("appointment"),
				Url:            lago.GetterRoutePath("appointments.UpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}),
				Method:         http.MethodPost,
				Title:          "Edit Appointment",
				Subtitle:       "Update appointment details",
				ChildrenInput:  appointmentFormFields(),
				ChildrenAction: []components.PageInterface{components.ButtonSubmit{Label: "Save Appointment"}},
			},
		},
	})
}

func registerTable() {
	lago.RegistryPage.Register("appointments.AppointmentTable", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentMenu"}},
		Children: []components.PageInterface{
			components.DataTable[Appointment]{
				UID:             "appointment-table",
				Data:            getters.GetterKey[components.ObjectList[Appointment]]("appointments"),
				Title:           "Appointments",
				Subtitle:        "List of appointments",
				CreateUrl:       lago.GetterRoutePath("appointments.CreateRoute", nil),
				OnClick:         getters.GetterNavigateGetter(lago.GetterRoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))})),
				FilterComponent: lago.DynamicPage{Name: "appointments.AppointmentFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey[string]("$row.Name")}}},
					{Label: "Location", Key: "Location", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey[string]("$row.Location")}}},
					{Label: "Phone", Key: "Phone", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey[string]("$row.Phone")}}},
					{Label: "Date & Time", Key: "Datetime", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.Datetime")}}},
					{Label: "Created By", Key: "CreatedBy", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterForeignKey[p_users.User, uint, string](getters.GetterKey[uint]("$row.CreatedByID"), "Name")}}},
					{Label: "Created At", Key: "CreatedAt", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.CreatedAt")}}},
				},
			},
		},
	})
}

func registerDetail() {
	generatedSection := []components.PageInterface{
		components.ContainerColumn{Classes: "mt-2 p-4 card card-body border rounded-box border-base-300", Children: []components.PageInterface{
			components.ContainerRow{Classes: "flex flex-wrap justify-between items-center gap-4 mb-4", Children: []components.PageInterface{
				components.FieldTitle{Getter: getters.GetterStatic("Generated Letter")},
				components.ContainerColumn{Classes: "flex gap-2", Children: []components.PageInterface{
					components.ButtonLink{Label: "Send via WhatsApp", Link: getters.GetterFormat("https://wa.me/%v?text=%v", getters.GetterAny(getters.GetterKey[string]("$in.Phone")), getters.GetterAny(getters.GetterQueryEscape(getters.GetterKey[string]("$in.GeneratedLetter")))), Classes: "btn-outline btn-success btn-sm"},
					components.ButtonModal{Label: "Edit with AI", Url: lago.GetterRoutePath("appointments.AiEditFormRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}), Classes: "btn-outline btn-secondary btn-sm"},
					components.ButtonPost{Label: "Regenerate Letter", URL: lago.GetterRoutePath("appointments.GenerateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}), Classes: "btn-outline btn-primary btn-sm"},
				}},
			}},
			components.FieldMarkdown{Getter: getters.GetterKey[string]("$in.GeneratedLetter")},
		}},
	}

	pendingSection := []components.PageInterface{
		components.ContainerRow{Classes: "flex gap-2 items-center", Children: []components.PageInterface{
			components.FieldText{Getter: getters.GetterStatic("Generating..."), Classes: "btn-primary"},
			components.ButtonPost{Label: "Cancel Generation", URL: lago.GetterRoutePath("appointments.CancelRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}), Classes: "btn-outline btn-error btn-sm"},
		}},
	}

	idleSection := []components.PageInterface{
		components.ButtonPost{Label: "Generate Letter with AI", URL: lago.GetterRoutePath("appointments.GenerateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}), Classes: "btn-primary"},
	}

	lago.RegistryPage.Register("appointments.AppointmentDetail", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentDetailMenu"}},
		Children: []components.PageInterface{
			components.Detail[Appointment]{
				Getter: getters.GetterKey[Appointment]("appointment"),
				Children: []components.PageInterface{
					components.ContainerColumn{Children: []components.PageInterface{
						components.ShowIf{Getter: getters.GetterAny(getters.GetterKey[bool]("OverlapWarning")), Children: []components.PageInterface{
							components.ContainerColumn{Classes: "bg-warning rounded-box border border-base-300 mb-4 shadow-sm gap-4 p-4", Children: []components.PageInterface{
								components.ContainerRow{Classes: "flex items-center gap-2 font-semibold", Children: []components.PageInterface{
									components.Icon{Name: "exclamation-triangle", Classes: "w-5 h-5"},
									components.FieldText{Getter: getters.GetterStatic("Overlapping Appointments:")},
								}},
								components.FieldList{
									Getter:  getters.GetterAny(getters.GetterKey[[]map[string]any]("OverlapWarningList")),
									Classes: "flex flex-col gap-2 pl-4",
									Children: []components.PageInterface{
										components.ContainerRow{Classes: "items-center gap-2", Children: []components.PageInterface{
											components.ButtonLink{GetterLabel: getters.GetterKey[string]("$row.Name"), Link: lago.GetterRoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))})},
											components.FieldText{Getter: getters.GetterStatic(" — ")},
											components.FieldText{Getter: getters.GetterKey[string]("$row.Date")},
										}},
									},
								},
							}},
						}},
						components.ContainerRow{Classes: "flex justify-between items-start", Children: []components.PageInterface{
							components.ContainerColumn{Children: []components.PageInterface{
								components.FieldTitle{Getter: getters.GetterKey[string]("$in.Name")},
								components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Location")},
							}},
						}},
						components.LabelInline{Title: "Phone", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey[string]("$in.Phone")}}},
						components.LabelInline{Title: "Date & Time", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$in.Datetime")}}},
						components.LabelInline{Title: "Remarks", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey[string]("$in.Remarks")}}},
						components.LabelInline{Title: "Extra Info", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey[string]("$in.ExtraInfo")}}},
						components.LabelInline{Title: "Created By", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterForeignKey[p_users.User, uint, string](getters.GetterKey[uint]("$in.CreatedByID"), "Name")}}},
						components.LabelInline{Title: "Created At", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$in.CreatedAt")}}},
						components.ContainerColumn{Classes: "mt-6", Children: []components.PageInterface{
							components.ShowIf{Getter: getters.GetterAny(getters.GetterKey[string]("$in.GeneratedLetter")), Children: generatedSection},
							components.ShowIf{Getter: getters.GetterAny(getters.GetterKey[bool]("GenerationPending")), Children: pendingSection},
							components.ShowIf{Getter: getters.GetterAny(getterIdleGeneration()), Children: idleSection},
						}},
					}},
				},
			},
		},
	})
}

func getterIdleGeneration() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		if content, _ := getters.IfOrGetter(getters.GetterKey[string]("$in.GeneratedLetter"), ctx, ""); content != "" {
			return false, nil
		}
		if pending, _ := getters.IfOrGetter(getters.GetterKey[bool]("GenerationPending"), ctx, false); pending {
			return false, nil
		}
		return true, nil
	}
}

func registerModal() {
	lago.RegistryPage.Register("appointments.AiEditModal", components.Modal{
		UID:   "ai-edit-modal",
		Title: "Edit with AI",
		Children: []components.PageInterface{
			components.FormComponent[Appointment]{
				Getter: getters.GetterKey[Appointment]("appointment"),
				Url:    lago.GetterRoutePath("appointments.AiEditRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("appointment.ID"))}),
				Method: http.MethodPost,
				ChildrenInput: []components.PageInterface{
					components.InputTextarea{Name: "generated_letter", Label: "Current Letter Content", Getter: getters.GetterKey[string]("$in.GeneratedLetter"), Rows: 8},
					components.InputTextarea{Name: "instructions", Label: "Instructions for AI", Rows: 4, Required: true},
				},
				ChildrenAction: []components.PageInterface{
					components.ContainerRow{Classes: "flex justify-end gap-2", Children: []components.PageInterface{
						components.ButtonSubmit{Label: "Generate", Classes: "btn-secondary"},
					}},
				},
			},
		},
	})
}

func registerDelete() {
	lago.RegistryPage.Register("appointments.AppointmentDeleteForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentDetailMenu"}},
		Children: []components.PageInterface{
			components.DeleteConfirmation{
				Title:     "Confirm Deletion",
				Message:   "Are you sure you want to delete this appointment?",
				CancelUrl: lago.GetterRoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("appointment.ID"))}),
			},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("appointments.AppointmentSelectionTable", components.Modal{
		UID:   "appointment-selection-modal",
		Title: "Select Appointment",
		Children: []components.PageInterface{
			components.DataTable[Appointment]{
				UID:             "appointment-selection-table",
				Data:            getters.GetterKey[components.ObjectList[Appointment]]("appointments"),
				OnClick:         getters.GetterSelect("appointment", getters.GetterKey[uint]("$row.ID"), getters.GetterKey[string]("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "appointments.AppointmentFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey[string]("$row.Name")}}},
					{Label: "Location", Key: "Location", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey[string]("$row.Location")}}},
					{Label: "Phone", Key: "Phone", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey[string]("$row.Phone")}}},
					{Label: "Date & Time", Key: "Datetime", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey[string]("$row.Datetime")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("appointments.AppointmentCardTimelineFilter", components.FormComponent[Appointment]{
		Url:    lago.GetterRoutePath("appointments.CardTimelineRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Date", Name: "Date", Getter: getters.GetterKey[string]("$get.Date")},
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
			components.Timeline[Appointment]{
				UID:             "appointment-timeline",
				Title:           "Appointments Timeline",
				Data:            getters.GetterKey[components.ObjectList[Appointment]]("appointments"),
				FilterComponent: lago.DynamicPage{Name: "appointments.AppointmentCardTimelineFilter"},
				OnClick:         lago.GetterRoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))}),
				Children: []components.PageInterface{
					components.ContainerColumn{Children: []components.PageInterface{
						components.ContainerRow{Classes: "flex justify-between items-start mb-2", Children: []components.PageInterface{
							components.FieldTitle{Getter: getters.GetterKey[string]("$row.Name")},
							components.FieldText{Getter: getters.GetterKey[string]("$row.Datetime"), Classes: "text-sm text-gray-500 font-medium whitespace-nowrap"},
						}},
						components.ContainerRow{Classes: "flex items-center gap-1 text-sm text-gray-600 mb-1", Children: []components.PageInterface{
							components.Icon{Name: "map-pin", Classes: "w-4 h-4"},
							components.FieldText{Getter: getters.GetterKey[string]("$row.Location")},
						}},
						components.ContainerRow{Classes: "flex items-center gap-1 text-sm text-gray-600 mb-2", Children: []components.PageInterface{
							components.Icon{Name: "phone", Classes: "w-4 h-4"},
							components.FieldText{Getter: getters.GetterKey[string]("$row.Phone")},
						}},
						components.ShowIf{Getter: getters.GetterAny(getters.GetterKey[string]("$row.Remarks")), Children: []components.PageInterface{
							components.FieldText{Getter: getters.GetterKey[string]("$row.Remarks"), Classes: "text-sm italic border-t pt-2 mt-2"},
						}},
					}},
				},
			},
		},
	})
}
