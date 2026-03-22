package p_semesters

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
	lago.RegistryPage.Register("semesters.SemesterMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Semesters"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Semesters"),
				Url:   lago.GetterRoutePath("semesters.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("semesters.SemesterDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Semester: %s", getters.GetterAny(getters.GetterKey[string]("semester.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Semesters"),
			Url:   lago.GetterRoutePath("semesters.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Semester Detail"),
				Url: lago.GetterRoutePath("semesters.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("semester.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Semester"),
				Url: lago.GetterRoutePath("semesters.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("semester.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Semester"),
				Url: lago.GetterRoutePath("semesters.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("semester.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("semesters.SemesterFilter", &components.FormComponent[Semester]{
		Url:    lago.GetterRoutePath("semesters.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.GetterKey[string]("$get.Code"),
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

	lago.RegistryPage.Register("semesters.SemesterSelectionFilter", &components.FormComponent[Semester]{
		Url:    lago.GetterRoutePath("semesters.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.GetterKey[string]("$get.Code"),
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
		return getters.GetterKey[string]("$in.Code")(ctx)
	}
}

func semesterFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "semesters.SemesterFormFieldsBody",
		},
		Children: []components.PageInterface{
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Name",
								Name:     "Name",
								Required: true,
								Getter:   getters.GetterKey[string]("$in.Name"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Code"),
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
						Error: getters.GetterKey[error]("$error.Start"),
						Children: []components.PageInterface{
							&components.InputDatetime{
								Label:    "Start Date & Time",
								Name:     "Start",
								Required: true,
								Getter:   getters.GetterKey[time.Time]("$in.Start"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.End"),
						Children: []components.PageInterface{
							&components.InputDatetime{
								Label:    "End Date & Time",
								Name:     "End",
								Required: true,
								Getter:   getters.GetterKey[time.Time]("$in.End"),
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.IsActive"),
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
	lago.RegistryPage.Register("semesters.SemesterFormFields", semesterFormFields())

	lago.RegistryPage.Register("semesters.SemesterCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "semesters.SemesterMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Semester]{
				Url:      lago.GetterRoutePath("semesters.CreateRoute", nil),
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

	lago.RegistryPage.Register("semesters.SemesterUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "semesters.SemesterDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Semester]{
				Getter: getters.GetterKey[Semester]("semester"),
				Url: lago.GetterRoutePath("semesters.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
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
	lago.RegistryPage.Register("semesters.SemesterTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "semesters.SemesterMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Semester]{
				Page:      components.Page{Key: "semesters.SemesterTableBody"},
				UID:       "semester-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[Semester]]("semesters"),
				CreateUrl: lago.GetterRoutePath("semesters.CreateRoute", nil),
				OnClick: getters.GetterNavigateGetter(
					lago.GetterRoutePath("semesters.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
					}),
				),
				FilterComponent: lago.DynamicPage{Name: "semesters.SemesterFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Code")},
						},
					},
					{
						Label: "Start",
						Name:  "Start",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.Start")},
						},
					},
					{
						Label: "End",
						Name:  "End",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.End")},
						},
					},
					{
						Label: "Active",
						Name:  "IsActive",
						Children: []components.PageInterface{
							&components.FieldCheckbox{Getter: getters.GetterKey[bool]("$row.IsActive")},
						},
					},
				},
			},
		},
	})
}

// --- Detail & Delete ---

func registerDetailPages() {
	lago.RegistryPage.Register("semesters.SemesterDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "semesters.SemesterDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Semester]{
				Getter: getters.GetterKey[Semester]("semester"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "semesters.SemesterDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Code")},
							&components.LabelInline{
								Title: "Active",
								Children: []components.PageInterface{
									&components.FieldCheckbox{Getter: getters.GetterKey[bool]("$in.IsActive")},
								},
							},
							&components.LabelInline{
								Title: "Start",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$in.Start")},
								},
							},
							&components.LabelInline{
								Title: "End",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$in.End")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("semesters.SemesterDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "semesters.SemesterDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this semester?",
				CancelUrl: lago.GetterRoutePath("semesters.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("semester.ID")),
				}),
			},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("semesters.SemesterSelectionTable", &components.Modal{
		UID:   "semester-selection-modal",
		Title: "Select Semester",
		Children: []components.PageInterface{
			&components.DataTable[Semester]{
				Page:            components.Page{Key: "semesters.SemesterSelectionTableBody"},
				UID:             "semester-selection-table",
				Data:            getters.GetterKey[components.ObjectList[Semester]]("semesters"),
				OnClick:         getters.GetterSelect("SemesterID", getters.GetterKey[uint]("$row.ID"), getters.GetterKey[string]("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "semesters.SemesterSelectionFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Code")},
						},
					},
					{
						Label: "Start",
						Name:  "Start",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.Start")},
						},
					},
					{
						Label: "Active",
						Name:  "IsActive",
						Children: []components.PageInterface{
							&components.FieldCheckbox{Getter: getters.GetterKey[bool]("$row.IsActive")},
						},
					},
				},
			},
		},
	})
}
