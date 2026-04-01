package p_nirmancampus_sessions

import (
	"context"
	"net/http"
	"time"

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
	lago.RegistryPage.Register("sessions.SemesterMenu", &components.SidebarMenu{
		Title: getters.Static("sessions"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All sessions"),
				Url:   lago.GetterRoutePath("sessions.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("sessions.SemesterDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Semester: %s", getters.Any(getters.Key[string]("semester.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all sessions"),
			Url:   lago.GetterRoutePath("sessions.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Semester Detail"),
				Url: lago.GetterRoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("semester.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit Semester"),
				Url: lago.GetterRoutePath("sessions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("semester.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Delete Semester"),
				Url: lago.GetterRoutePath("sessions.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("semester.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("sessions.SemesterFilter", &components.FormComponent[Semester]{
		Url:    lago.GetterRoutePath("sessions.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.Key[string]("$get.Code"),
			},
			&components.InputTernary{
				Label:      "Active",
				Name:       "IsActiveFilter",
				TrueLabel:  "Active Only",
				FalseLabel: "Inactive Only",
				NoneLabel:  "All",
				// Intentionally omit Getter: we want the default selection to be "All".
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply Filters"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})

	lago.RegistryPage.Register("sessions.sessionselectionFilter", &components.FormComponent[Semester]{
		Url:    lago.GetterRoutePath("sessions.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.Key[string]("$get.Code"),
			},
			&components.InputTernary{
				Label:      "Active",
				Name:       "IsActiveFilter",
				TrueLabel:  "Active Only",
				FalseLabel: "Inactive Only",
				NoneLabel:  "All",
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})
}

// --- Form Fields ---

func isActiveGetter() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		in := ctx.Value(getters.ContextKeyIn)
		if in == nil {
			// Create forms don't provide $in; default the checkbox to "true".
			return true, nil
		}
		m, ok := in.(map[string]any)
		if !ok {
			return true, nil
		}
		raw, ok := m["IsActive"]
		if !ok || raw == nil {
			return true, nil
		}
		v, ok := raw.(bool)
		if !ok {
			return true, nil
		}
		return v, nil
	}
}

func semesterCodeInputGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		// Code is auto-generated in BeforeSave when blank, so keep it empty on create.
		return getters.Key[string]("$in.Code")(ctx)
	}
}

func semesterFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "sessions.SemesterFormFieldsBody",
		},
		Children: []components.PageInterface{
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Name",
								Name:     "Name",
								Required: true,
								Getter:   getters.Key[string]("$in.Name"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Code"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Code",
								Name:   "Code",
								Getter: semesterCodeInputGetter(),
							},
						},
					},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Start"),
						Children: []components.PageInterface{
							&components.InputDatetime{
								Label:    "Start Date & Time",
								Name:     "Start",
								Required: true,
								Getter:   getters.Key[time.Time]("$in.Start"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.End"),
						Children: []components.PageInterface{
							&components.InputDatetime{
								Label:    "End Date & Time",
								Name:     "End",
								Required: true,
								Getter:   getters.Key[time.Time]("$in.End"),
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.IsActive"),
				Children: []components.PageInterface{
					&components.InputCheckbox{
						Label:    "Active",
						Name:     "IsActive",
						Getter:   isActiveGetter(),
						Required: false,
					},
				},
			},
		},
	}
}

// --- Form Pages ---

func registerFormPages() {
	lago.RegistryPage.Register("sessions.SemesterFormFields", semesterFormFields())

	lago.RegistryPage.Register("sessions.SemesterCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.SemesterMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Semester]{
				Url:      lago.GetterRoutePath("sessions.CreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create Semester",
				Subtitle: "Create a new semester",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					semesterFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Semester"},
				},
			},
		},
	})

	lago.RegistryPage.Register("sessions.SemesterUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.SemesterDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Semester]{
				Getter: getters.Key[Semester]("semester"),
				Url: lago.GetterRoutePath("sessions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Semester",
				Subtitle: "Update semester details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					semesterFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Semester"},
				},
			},
		},
	})
}

// --- Tables ---

func registerTablePages() {
	lago.RegistryPage.Register("sessions.SemesterTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.SemesterMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Semester]{
				Page:      components.Page{Key: "sessions.SemesterTableBody"},
				UID:       "semester-table",
				Classes:   "w-full",
				Data:      getters.Key[components.ObjectList[Semester]]("sessions"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "sessions.SemesterFilter"}},
					&components.TableButtonCreate{Link: lago.GetterRoutePath("sessions.CreateRoute", nil)},
				},
				OnClick: getters.NavigateGetter(
					lago.GetterRoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Code")},
						},
					},
					{
						Label: "Start",
						Name:  "Start",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Start")},
						},
					},
					{
						Label: "End",
						Name:  "End",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.End")},
						},
					},
					{
						Label: "Active",
						Name:  "IsActive",
						Children: []components.PageInterface{
							&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsActive")},
						},
					},
				},
			},
		},
	})
}

// --- Detail & Delete ---

func registerDetailPages() {
	lago.RegistryPage.Register("sessions.SemesterDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.SemesterDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Semester]{
				Getter: getters.Key[Semester]("semester"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "sessions.SemesterDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Code")},
							&components.LabelInline{
								Title: "Active",
								Children: []components.PageInterface{
									&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsActive")},
								},
							},
							&components.LabelInline{
								Title: "Start",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.Start")},
								},
							},
							&components.LabelInline{
								Title: "End",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.End")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("sessions.SemesterDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.SemesterDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this semester?",
				CancelUrl: lago.GetterRoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("semester.ID")),
				}),
			},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("sessions.sessionselectionTable", &components.Modal{
		UID:   "semester-selection-modal",
		Title: "Select Semester",
		Children: []components.PageInterface{
			&components.DataTable[Semester]{
				Page:            components.Page{Key: "sessions.sessionselectionTableBody"},
				UID:             "semester-selection-table",
				Data:            getters.Key[components.ObjectList[Semester]]("sessions"),
				OnClick: getters.Select("SemesterID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "sessions.sessionselectionFilter"}},
				},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Code")},
						},
					},
					{
						Label: "Start",
						Name:  "Start",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Start")},
						},
					},
					{
						Label: "Active",
						Name:  "IsActive",
						Children: []components.PageInterface{
							&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsActive")},
						},
					},
				},
			},
		},
	})
}
