package p_totschool_appointments

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
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
			components.SidebarMenuItem{Title: getters.Static("Delete Appointment"), Url: lago.RoutePath("appointments.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("appointment.ID"))})},
		},
	})
}

func registerFilter() {
	lago.RegistryPage.Register("appointments.AppointmentFilter", components.FormComponent[Appointment]{
		Attr: getters.FormAttr(http.MethodGet, getters.FormSubmitGet(lago.RoutePath("appointments.ListRoute", nil))),

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
			components.FormComponent[Appointment]{
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmit(lago.RoutePath("appointments.CreateRoute", nil))),

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
				Getter:         getters.Key[Appointment]("appointment"),
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmit(lago.RoutePath("appointments.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}))),

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

func registerDetail() {
	generatedSection := []components.PageInterface{
		components.ContainerColumn{Classes: "mt-2 p-4 card card-body border rounded-box border-base-300", Children: []components.PageInterface{
			components.ContainerRow{Classes: "flex flex-wrap justify-between items-center gap-4 mb-4", Children: []components.PageInterface{
				components.FieldTitle{Getter: getters.Static("Generated Letter")},
				components.ContainerColumn{Classes: "flex gap-2", Children: []components.PageInterface{
					components.ButtonLink{Classes: "btn-outline btn-success btn-sm", Label: "Send via WhatsApp", Link: getters.Format("https://wa.me/%v?text=%v", getters.Any(getters.Key[string]("$in.Phone")), getters.Any(getters.QueryEscape(getters.Key[string]("$in.GeneratedLetter"))))},
					components.ButtonModal{Label: "Edit with AI", Url: lago.RoutePath("appointments.AiEditFormRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}), Classes: "btn-outline btn-secondary btn-sm"},
					components.ButtonPost{Label: "Regenerate Letter", URL: lago.RoutePath("appointments.GenerateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}), Classes: "btn-outline btn-primary btn-sm"},
				}},
			}},
			components.FieldMarkdown{Getter: getters.Key[string]("$in.GeneratedLetter")},
		}},
	}

	pendingSection := []components.PageInterface{
		components.HTMXPolling{
			URL: lago.RoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}),
			Children: []components.PageInterface{
				components.ContainerRow{Classes: "flex gap-2 items-center", Children: []components.PageInterface{
					components.FieldText{Getter: getters.Static("Generating...")},
					components.ButtonPost{
						Label:   "Cancel Generation",
						URL:     lago.RoutePath("appointments.CancelRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}),
						Classes: "btn-outline btn-error btn-sm",
					},
				}},
			},
		},
	}

	idleSection := []components.PageInterface{
		components.ButtonPost{Label: "Generate Letter with AI", URL: lago.RoutePath("appointments.GenerateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}), Classes: "btn-primary"},
	}

	lago.RegistryPage.Register("appointments.AppointmentDetail", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentDetailMenu"}},
		Children: []components.PageInterface{
			components.Detail[Appointment]{
				Getter: getters.Key[Appointment]("appointment"),
				Children: []components.PageInterface{
					components.ContainerColumn{Children: []components.PageInterface{
						components.ShowIf{Getter: getters.Any(getters.Key[bool]("OverlapWarning")), Children: []components.PageInterface{
							components.ContainerColumn{Classes: "bg-warning rounded-box border border-base-300 mb-4 shadow-sm gap-4 p-4", Children: []components.PageInterface{
								components.ContainerRow{Classes: "flex items-center gap-2 font-semibold", Children: []components.PageInterface{
									components.Icon{Name: "exclamation-triangle", Classes: "w-5 h-5"},
									components.FieldText{Getter: getters.Static("Overlapping Appointments:")},
								}},
								components.FieldList{
									Getter:  getters.Any(getters.Key[[]map[string]any]("OverlapWarningList")),
									Classes: "flex flex-col gap-2 pl-4",
									Children: []components.PageInterface{
										components.ContainerRow{Classes: "items-center gap-2", Children: []components.PageInterface{
											components.ButtonLink{GetterLabel: getters.Key[string]("$row.Name"), Link: lago.RoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})},
											components.FieldText{Getter: getters.Static(" — ")},
											components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Date")},
										}},
									},
								},
							}},
						}},
						components.ContainerRow{Classes: "flex justify-between items-start", Children: []components.PageInterface{
							components.ContainerColumn{Children: []components.PageInterface{
								components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
								components.FieldSubtitle{Getter: getters.Key[string]("$in.Location")},
							}},
						}},
						components.LabelInline{Title: "Phone", Children: []components.PageInterface{components.FieldPhone{Getter: getters.Key[string]("$in.Phone")}}},
						components.LabelInline{Title: "Date & Time", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.Key[time.Time]("$in.Datetime")}}},
						components.LabelInline{Title: "Remarks", Children: []components.PageInterface{components.FieldText{Getter: getters.Key[string]("$in.Remarks")}}},
						components.LabelInline{Title: "Extra Info", Children: []components.PageInterface{components.FieldText{Getter: getters.Key[string]("$in.ExtraInfo")}}},
						components.LabelInline{Title: "Created By", Children: []components.PageInterface{components.FieldText{Getter: getters.ForeignKey[p_users.User, uint, string](getters.Key[uint]("$in.CreatedByID"), "Name")}}},
						components.LabelInline{Title: "Created At", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.Key[time.Time]("$in.CreatedAt")}}},
						components.ContainerColumn{Classes: "mt-6", Children: []components.PageInterface{
							components.ShowIf{Getter: getters.Any(getterGenerated()), Children: generatedSection},
							components.ShowIf{Getter: getters.Any(getterGenerationPending()), Children: pendingSection},
							components.ShowIf{Getter: getters.Any(getterIdleGeneration()), Children: idleSection},
						}},
					}},
				},
			},
		},
	})
}

func getterGenerated() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		id, err := getters.Key[*int]("$in.GenerationID")(ctx)
		if err != nil {
			slog.Error("Error while getting id for checking if appointment is idle", "error", err)
			return false, err
		}
		content, err := getters.Key[string]("$in.GeneratedLetter")(ctx)
		if err != nil {
			slog.Error("Error while getting content for checking if appointment is idle", "error", err)
			return false, err
		}
		if id == nil && content != "" {
			return true, nil
		}
		return false, nil
	}
}

func getterGenerationPending() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		id, err := getters.Key[*int]("$in.GenerationID")(ctx)
		if err != nil {
			slog.Error("Error while getting id for checking if appointment is idle", "error", err)
			return false, err
		}
		content, err := getters.Key[string]("$in.GeneratedLetter")(ctx)
		if err != nil {
			slog.Error("Error while getting content for checking if appointment is idle", "error", err)
			return false, err
		}
		if id != nil && content == "" {
			return true, nil
		}
		return false, nil
	}
}

func getterIdleGeneration() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		id, err := getters.Key[*int]("$in.GenerationID")(ctx)
		if err != nil {
			slog.Error("Error while getting id for checking if appointment is idle", "error", err)
			return false, err
		}
		content, err := getters.Key[string]("$in.GeneratedLetter")(ctx)
		if err != nil {
			slog.Error("Error while getting content for checking if appointment is idle", "error", err)
			return false, err
		}
		if id == nil && content == "" {
			return true, nil
		}
		return false, nil
	}
}

func registerModal() {
	lago.RegistryPage.Register("appointments.AiEditModal", components.Modal{
		UID: "ai-edit-modal",
		Children: []components.PageInterface{
			components.FormComponent[Appointment]{
				Getter:   getters.Key[Appointment]("appointment"),
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmitCloseModal(lago.RoutePath("appointments.AiEditRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("appointment.ID"))}))),

				Title:    "Edit with AI",
				ChildrenInput: []components.PageInterface{
					components.ContainerError{
						Error: getters.Key[error]("$error.generated_letter"),
						Children: []components.PageInterface{
							components.InputTextarea{Name: "generated_letter", Label: "Current Letter Content", Getter: getters.Key[string]("$in.GeneratedLetter"), Rows: 8},
						},
					},
					components.ContainerError{
						Error: getters.Key[error]("$error.instructions"),
						Children: []components.PageInterface{
							components.InputTextarea{Name: "instructions", Label: "Instructions for AI", Rows: 4, Required: true},
						},
					},
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
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this appointment?",
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmit(lago.RoutePath("appointments.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("appointment.ID")),
				}))),
			},
		},
	})
}

// --- Selection Tables ---

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
		Attr: getters.FormAttr(http.MethodGet, getters.FormSubmitGet(lago.RoutePath("appointments.CardTimelineRoute", nil))),

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
