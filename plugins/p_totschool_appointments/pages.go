package p_totschool_appointments

import (
	"context"
	"net/http"

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
			Url:   getters.GetterStatic("/apps/"),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{Title: getters.GetterStatic("All Appointments"), Url: lago.RoutePathGetter("appointments.ListRoute")},
			components.SidebarMenuItem{Title: getters.GetterStatic("Appointments Timeline"), Url: lago.RoutePathGetter("appointments.CardTimelineRoute")},
			components.SidebarMenuItem{Title: getters.GetterStatic("Create Appointment"), Url: lago.RoutePathGetter("appointments.CreateRoute")},
		},
	})

	lago.RegistryPage.Register("appointments.AppointmentDetailMenu", components.SidebarMenu{
		Title: getters.GetterFormat("Appointment: %s", getters.GetterKey("appointment.Name")),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Appointments"),
			Url:   lago.RoutePathGetter("appointments.ListRoute"),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{Title: getters.GetterStatic("Appointment Detail"), Url: getters.GetterFormat(AppUrl+"%v/", getters.GetterKey("appointment.ID"))},
			components.SidebarMenuItem{Title: getters.GetterStatic("Edit Appointment"), Url: getters.GetterFormat(AppUrl+"%v/edit/", getters.GetterKey("appointment.ID"))},
			components.SidebarMenuItem{Title: getters.GetterStatic("Delete Appointment"), Url: getters.GetterFormat(AppUrl+"%v/delete/", getters.GetterKey("appointment.ID"))},
		},
	})
}

func registerFilter() {
	lago.RegistryPage.Register("appointments.AppointmentFilter", components.FormComponent{
		Url:    getters.GetterStatic(AppUrl),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "name", Getter: getters.GetterKey("$get.name")},
			components.InputText{Label: "Location", Name: "location", Getter: getters.GetterKey("$get.location")},
			components.InputText{Label: "Date", Name: "date", Getter: getters.GetterKey("$get.date")},
			components.InputManyToMany{
				Label:       "Created By",
				Name:        "created_by",
				Url:         getters.GetterStatic(p_users.AppUrl + "multi-select/"),
				DisplayAttr: "Name",
				Placeholder: "Select users...",
				Getter:      getters.GetterKey("$get.created_by"),
			},
			components.InputCheckbox{Label: "Overlaps Only", Name: "overlapping", Getter: getters.GetterKey("$get.overlapping")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply Filters"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})
}

func appointmentFormFields() []components.PageInterface {
	return []components.PageInterface{
		components.InputText{Label: "Name", Name: "name", Required: true, Getter: getters.GetterKey("$in.Name")},
		components.InputTextarea{Label: "Location", Name: "location", Required: true, Getter: getters.GetterKey("$in.Location"), Rows: 2},
		components.ContainerRow{Classes: "grid grid-cols-1 gap-1 md:grid-cols-2", Children: []components.PageInterface{
			components.InputPhone{Label: "Phone", Name: "phone", Required: true, Getter: getters.GetterKey("$in.Phone")},
			components.InputDatetime{Label: "Date & Time", Name: "datetime", Required: true, Getter: getters.GetterKey("$in.Datetime")},
		}},
		components.InputTextarea{Label: "Remarks", Name: "remarks", Getter: getters.GetterKey("$in.Remarks"), Rows: 2},
		components.InputTextarea{Label: "Extra Info (For AI)", Name: "extra_info", Getter: getters.GetterKey("$in.ExtraInfo"), Rows: 2},
	}
}

func registerForms() {
	lago.RegistryPage.Register("appointments.AppointmentCreateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentMenu"}},
		Children: []components.PageInterface{
			components.FormComponent{
				Url:            getters.GetterStatic(AppUrl + "create/"),
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
			components.FormComponent{
				Getter:         getters.GetterKey("appointment"),
				Url:            getters.GetterFormat(AppUrl+"%v/edit/", getters.GetterKey("$in.ID")),
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
			components.DataTable{
				UID:             "appointment-table",
				Data:            getters.GetterKey("appointments"),
				Title:           "Appointments",
				Subtitle:        "List of appointments",
				CreateUrl:       lago.RoutePathGetter("appointments.CreateRoute"),
				OnClick:         getters.GetterNavigate(AppUrl+"%v/", getters.GetterKey("$row.ID")),
				FilterComponent: lago.DynamicPage{Name: "appointments.AppointmentFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.Name")}}},
					{Label: "Location", Key: "Location", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.Location")}}},
					{Label: "Phone", Key: "Phone", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.Phone")}}},
					{Label: "Date & Time", Key: "Datetime", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.Datetime")}}},
					{Label: "Created By", Key: "CreatedBy", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterForeignKey[p_users.User](getters.GetterKey("$row.CreatedByID"), "Name")}}},
					{Label: "Created At", Key: "CreatedAt", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.CreatedAt")}}},
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
					components.ButtonLink{Label: "Send via WhatsApp", Link: getters.GetterFormat("https://wa.me/%v?text=%v", getters.GetterKey("$in.Phone"), getters.GetterQueryEscape(getters.GetterKey("$in.GeneratedLetter"))), Classes: "btn-outline btn-success btn-sm"},
					components.ButtonModal{Label: "Edit with AI", Url: getters.GetterFormat(AppUrl+"%v/ai-edit/form/", getters.GetterKey("$in.ID")), Classes: "btn-outline btn-secondary btn-sm"},
					components.ButtonPost{Label: "Regenerate Letter", Url: getters.GetterFormat(AppUrl+"%v/generate/", getters.GetterKey("$in.ID")), Classes: "btn-outline btn-primary btn-sm"},
				}},
			}},
			components.FieldMarkdown{Getter: getters.GetterKey("$in.GeneratedLetter"), Classes: "bg-base-100 p-8 rounded-lg shadow border whitespace-pre-wrap"},
		}},
	}

	pendingSection := []components.PageInterface{
		components.ContainerRow{Classes: "flex gap-2 items-center", Children: []components.PageInterface{
			components.FieldText{Getter: getters.GetterStatic("Generating..."), Classes: "btn-primary"},
			components.ButtonPost{Label: "Cancel Generation", Url: getters.GetterFormat(AppUrl+"%v/cancel/", getters.GetterKey("$in.ID")), Classes: "btn-outline btn-error btn-sm"},
		}},
	}

	idleSection := []components.PageInterface{
		components.ButtonPost{Label: "Generate Letter with AI", Url: getters.GetterFormat(AppUrl+"%v/generate/", getters.GetterKey("$in.ID")), Classes: "btn-primary"},
	}

	lago.RegistryPage.Register("appointments.AppointmentDetail", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentDetailMenu"}},
		Children: []components.PageInterface{
			components.Detail{
				Getter: getters.GetterKey("appointment"),
				Children: []components.PageInterface{
					components.ContainerColumn{Children: []components.PageInterface{
						components.ShowIf{Getter: getters.GetterKey("overlap_warning"), Children: []components.PageInterface{
							components.ContainerColumn{Classes: "bg-warning rounded-box border border-base-300 mb-4 shadow-sm gap-4 p-4", Children: []components.PageInterface{
								components.ContainerRow{Classes: "flex items-center gap-2 font-semibold", Children: []components.PageInterface{
									components.Icon{Name: "exclamation-triangle", Classes: "w-5 h-5"},
									components.FieldText{Getter: getters.GetterStatic("Overlapping Appointments:")},
								}},
								components.FieldList{
									Getter:  getters.GetterKey("overlap_warning_list"),
									Classes: "flex flex-col gap-2 pl-4",
									Children: []components.PageInterface{
										components.ContainerRow{Classes: "flex items-center gap-2", Children: []components.PageInterface{
											components.ButtonLink{LabelGetter: getters.GetterKey("$row.Name"), Link: getters.GetterNavigate(AppUrl+"%v/", getters.GetterKey("$row.ID")), Classes: "link link-primary font-medium"},
											components.FieldText{Getter: getters.GetterStatic(" — ")},
											components.FieldText{Getter: getters.GetterKey("$row.Date")},
										}},
									},
								},
							}},
						}},
						components.ContainerRow{Classes: "flex justify-between items-start", Children: []components.PageInterface{
							components.ContainerColumn{Children: []components.PageInterface{
								components.FieldTitle{Getter: getters.GetterKey("$in.Name")},
								components.FieldSubtitle{Getter: getters.GetterKey("$in.Location")},
							}},
						}},
						components.LabelInline{Title: "Phone", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$in.Phone")}}},
						components.LabelInline{Title: "Date & Time", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$in.Datetime")}}},
						components.LabelInline{Title: "Remarks", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$in.Remarks")}}},
						components.LabelInline{Title: "Extra Info", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$in.ExtraInfo")}}},
						components.LabelInline{Title: "Created By", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterForeignKey[p_users.User](getters.GetterKey("$in.CreatedByID"), "Name")}}},
						components.LabelInline{Title: "Created At", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$in.CreatedAt")}}},
						components.ContainerColumn{Classes: "mt-6", Children: []components.PageInterface{
							components.ShowIf{Getter: getters.GetterKey("$in.GeneratedLetter"), Children: generatedSection},
							components.ShowIf{Getter: getters.GetterKey("generation_pending"), Children: pendingSection},
							components.ShowIf{Getter: getterIdleGeneration(), Children: idleSection},
						}},
					}},
				},
			},
		},
	})
}

func getterIdleGeneration() getters.Getter {
	return func(ctx context.Context) any {
		if content, _ := getters.IfOrGetter(getters.GetterKey("$in.GeneratedLetter"), ctx, "").(string); content != "" {
			return false
		}
		if getters.IfOrGetter(getters.GetterKey("generation_pending"), ctx, nil) != nil {
			return false
		}
		return true
	}
}

func registerModal() {
	lago.RegistryPage.Register("appointments.AiEditModal", components.Modal{
		UID:   "ai-edit-modal",
		Title: "Edit with AI",
		Children: []components.PageInterface{
			components.FormComponent{
				Getter: getters.GetterKey("appointment"),
				Url:    getters.GetterFormat(AppUrl+"%v/ai-edit/", getters.GetterKey("appointment.ID")),
				Method: http.MethodPost,
				ChildrenInput: []components.PageInterface{
					components.InputTextarea{Name: "generated_letter", Label: "Current Letter Content", Getter: getters.GetterKey("$in.GeneratedLetter"), Rows: 8},
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
				CancelUrl: getters.GetterFormat(AppUrl+"%v/", getters.GetterKey("appointment.ID")),
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
			components.DataTable{
				UID:             "appointment-selection-table",
				Data:            getters.GetterKey("appointments"),
				OnClick:         getters.GetterSelect("appointment", getters.GetterKey("$row.ID"), getters.GetterKey("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "appointments.AppointmentFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.Name")}}},
					{Label: "Location", Key: "Location", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.Location")}}},
					{Label: "Phone", Key: "Phone", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.Phone")}}},
					{Label: "Date & Time", Key: "Datetime", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.Datetime")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("appointments.TemplateSelectionTable", components.Modal{
		UID:   "template-selection-modal",
		Title: "Select Template",
		Children: []components.PageInterface{
			components.DataTable{
				UID:             "template-selection-table",
				Data:            getters.GetterKey("templates"),
				OnClick:         getters.GetterSelect("template", getters.GetterKey("$row.ID"), getters.GetterKey("$row.Name")),
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.Name")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("appointments.AppointmentCardTimelineFilter", components.FormComponent{
		Url:    lago.RoutePathGetter("appointments.CardTimelineRoute"),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Date", Name: "date", Getter: getters.GetterKey("$get.date")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply Filters"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("appointments.AppointmentCardTimeline", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "appointments.AppointmentMenu"}},
		Children: []components.PageInterface{
			components.Timeline{
				UID:             "appointment-timeline",
				Title:           "Appointments Timeline",
				Data:            getters.GetterKey("appointments"),
				FilterComponent: lago.DynamicPage{Name: "appointments.AppointmentCardTimelineFilter"},
				OnClick:         getters.GetterFormat(AppUrl+"%v/", getters.GetterKey("$row.ID")),
				Children: []components.PageInterface{
					components.ContainerColumn{Children: []components.PageInterface{
						components.ContainerRow{Classes: "flex justify-between items-start mb-2", Children: []components.PageInterface{
							components.FieldTitle{Getter: getters.GetterKey("$row.Name")},
							components.FieldText{Getter: getters.GetterKey("$row.Datetime"), Classes: "text-sm text-gray-500 font-medium whitespace-nowrap"},
						}},
						components.ContainerRow{Classes: "flex items-center gap-1 text-sm text-gray-600 mb-1", Children: []components.PageInterface{
							components.Icon{Name: "map-pin", Classes: "w-4 h-4"},
							components.FieldText{Getter: getters.GetterKey("$row.Location")},
						}},
						components.ContainerRow{Classes: "flex items-center gap-1 text-sm text-gray-600 mb-2", Children: []components.PageInterface{
							components.Icon{Name: "phone", Classes: "w-4 h-4"},
							components.FieldText{Getter: getters.GetterKey("$row.Phone")},
						}},
						components.ShowIf{Getter: getters.GetterKey("$row.Remarks"), Children: []components.PageInterface{
							components.FieldText{Getter: getters.GetterKey("$row.Remarks"), Classes: "text-sm italic border-t pt-2 mt-2"},
						}},
					}},
				},
			},
		},
	})
}


