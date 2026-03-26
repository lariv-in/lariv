package p_academicrecords

import (
	"context"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
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
	lago.RegistryPage.Register("academicrecords.AcademicRecordMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Academic Records"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Academic Records"),
				Url:   lago.GetterRoutePath("academicrecords.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("academicrecords.AcademicRecordDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Record: %s", getters.GetterAny(getters.GetterKey[string]("academicrecord.Student.User.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Academic Records"),
			Url:   lago.GetterRoutePath("academicrecords.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Academic Record Detail"),
				Url: lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("academicrecord.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Academic Record"),
				Url: lago.GetterRoutePath("academicrecords.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("academicrecord.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Academic Record"),
				Url: lago.GetterRoutePath("academicrecords.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("academicrecord.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordFilter", &components.FormComponent[AcademicRecord]{
		Url:    lago.GetterRoutePath("academicrecords.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Status",
				Name:   "Status",
				Getter: getters.GetterKey[string]("$get.Status"),
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
}

// --- Form Fields ---

func academicRecordFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "academicrecords.AcademicRecordFormFieldsBody",
		},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.StudentID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_nirmancampus_students.Student]{
								Label:       "Student",
								Name:        "StudentID",
								Required:    true,
								Url:         lago.GetterRoutePath("students.SelectRoute", nil),
								Display:     getters.GetterKey[string]("$in.StudentNo"),
								Placeholder: "Select a student...",
								Getter: getters.GetterAssociation[p_nirmancampus_students.Student](
									getters.GetterKey[uint]("$in.StudentID"),
								),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Status"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Status",
								Name:     "Status",
								Required: true,
								Getter:   getters.GetterKey[string]("$in.Status"),
							},
						},
					},
				},
			},
		},
	}
}

// --- Form Pages ---

func registerFormPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordFormFields", academicRecordFormFields())

	lago.RegistryPage.Register("academicrecords.AcademicRecordCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[AcademicRecord]{
				Url:    lago.GetterRoutePath("academicrecords.CreateRoute", nil),
				Method: http.MethodPost,
				Title:  "Create Academic Record",
				// Keep subtitle aligned with other apps.
				Subtitle: "Create a new academic record",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					// Embed directly so the companion plugin can patch by Page.Key.
					academicRecordFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Academic Record"},
				},
			},
		},
	})

	lago.RegistryPage.Register("academicrecords.AcademicRecordUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[AcademicRecord]{
				Getter: getters.GetterKey[AcademicRecord]("academicrecord"),
				Url: lago.GetterRoutePath("academicrecords.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Academic Record",
				Subtitle: "Update academic record details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					academicRecordFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Academic Record"},
				},
			},
		},
	})
}

// --- Tables ---

func registerTablePages() {
	createURLGetter := lago.GetterRoutePath("academicrecords.CreateRoute", nil)
	roleGetter := getters.GetterKey[string]("$role")

	lago.RegistryPage.Register("academicrecords.AcademicRecordTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[AcademicRecord]{
				Page:      components.Page{Key: "academicrecords.AcademicRecordTableBody"},
				UID:       "academicrecords-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[AcademicRecord]]("academicrecords"),
				CreateUrl: func(ctx context.Context) (string, error) {
					role, err := roleGetter(ctx)
					if err != nil {
						return "", err
					}
					if role != "superuser" && role != "admin" {
						return "", nil
					}
					return createURLGetter(ctx)
				},
				OnClick: getters.GetterNavigateGetter(
					lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
					}),
				),
				FilterComponent: lago.DynamicPage{Name: "academicrecords.AcademicRecordFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Student",
						Name:  "Student.User.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Student.User.Name")},
						},
					},
					{
						Label: "Status",
						Name:  "Status",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Status")},
						},
					},
				},
			},
		},
	})
}

// --- Detail & Delete ---

func registerDetailPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[AcademicRecord]{
				Getter: getters.GetterKey[AcademicRecord]("academicrecord"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "academicrecords.AcademicRecordDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Student.User.Name")},
							&components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Student.StudentNo")},
							&components.LabelInline{
								Title: "Status",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Status")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("academicrecords.AcademicRecordDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this academic record?",
				CancelUrl: lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("academicrecord.ID")),
				}),
			},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordSelectionTable", &components.Modal{
		UID:   "academicrecords-selection-modal",
		Title: "Select Academic Record",
		Children: []components.PageInterface{
			&components.DataTable[AcademicRecord]{
				Page:    components.Page{Key: "academicrecords.AcademicRecordSelectionTableBody"},
				UID:     "academicrecords-selection-table",
				Data:    getters.GetterKey[components.ObjectList[AcademicRecord]]("academicrecords"),
				OnClick: getters.GetterSelect("AcademicRecordID", getters.GetterKey[uint]("$row.ID"), getters.GetterKey[string]("$row.Status")),
				Columns: []components.TableColumn{
					{
						Label: "Student",
						Name:  "Student.User.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Student.User.Name")},
						},
					},
					{
						Label: "Status",
						Name:  "Status",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Status")},
						},
					},
				},
			},
		},
	})
}
