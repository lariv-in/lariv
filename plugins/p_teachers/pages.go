package p_teachers

import (
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
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
	lago.RegistryPage.Register("teachers.TeacherMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Teachers"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Teachers"),
				Url:   lago.GetterRoutePath("teachers.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("teachers.TeacherDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Teacher: %s", getters.GetterAny(getters.GetterKey[string]("teacher.User.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Teachers"),
			Url:   lago.GetterRoutePath("teachers.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Teacher Detail"),
				Url: lago.GetterRoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("teacher.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Teacher"),
				Url: lago.GetterRoutePath("teachers.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("teacher.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Teacher"),
				Url: lago.GetterRoutePath("teachers.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("teacher.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("teachers.TeacherFilter", &components.FormComponent[Teacher]{
		Url:    lago.GetterRoutePath("teachers.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Teacher Code",
				Name:   "Code",
				Getter: getters.GetterKey[string]("$get.Code"),
			},
			&components.InputText{
				Label:  "Name",
				Name:   "User.Name",
				Getter: getters.GetterKey[string]("$get.User.Name"),
			},
			&components.InputText{
				Label:  "Qualifications",
				Name:   "Qualifications",
				Getter: getters.GetterKey[string]("$get.Qualifications"),
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

	lago.RegistryPage.Register("teachers.TeacherSelectionFilter", &components.FormComponent[Teacher]{
		Url:    lago.GetterRoutePath("teachers.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "User.Name",
				Getter: getters.GetterKey[string]("$get.User.Name"),
			},
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.GetterKey[string]("$get.Code"),
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

	lago.RegistryPage.Register("teachers.TeacherMultiSelectionFilter", &components.FormComponent[Teacher]{
		Url:    lago.GetterRoutePath("teachers.MultiSelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "User.Name",
				Getter: getters.GetterKey[string]("$get.User.Name"),
			},
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.GetterKey[string]("$get.Code"),
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

// --- Form Fields & Forms ---

func teacherFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "teachers.TeacherFormFieldsBody",
		},
		Children: []components.PageInterface{
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.UserID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_users.User]{
								Label:       "User Account",
								Name:        "UserID",
								Required:    true,
								Getter:      getters.GetterAssociation[p_users.User](getters.GetterKey[uint]("$in.UserID")),
								Url:         lago.GetterRoutePath("users.SelectRoute", nil),
								Display:     getters.GetterKey[string]("$in.Name"),
								Placeholder: "Select a user...",
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Code"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Teacher Code",
								Name:     "Code",
								Required: true,
								Getter:   getters.GetterKey[string]("$in.Code"),
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.Qualifications"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:  "Qualifications",
						Name:   "Qualifications",
						Rows:   3,
						Getter: getters.GetterDeref(getters.GetterKey[*string]("$in.Qualifications")),
					},
				},
			},
		},
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("teachers.TeacherFormFields", teacherFormFields())

	lago.RegistryPage.Register("teachers.TeacherCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "teachers.TeacherMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Teacher]{
				Url:      lago.GetterRoutePath("teachers.CreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create Teacher",
				Subtitle: "Create a new teacher",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					teacherFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Teacher"},
				},
			},
		},
	})

	lago.RegistryPage.Register("teachers.TeacherUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "teachers.TeacherDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Teacher]{
				Getter:   getters.GetterKey[Teacher]("teacher"),
				Url:      lago.GetterRoutePath("teachers.UpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}),
				Method:   http.MethodPost,
				Title:    "Edit Teacher",
				Subtitle: "Update teacher details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					teacherFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Teacher"},
				},
			},
		},
	})
}

// --- Tables ---

func registerTablePages() {
	lago.RegistryPage.Register("teachers.TeacherTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "teachers.TeacherMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Teacher]{
				UID:             "teacher-table",
				Classes:         "w-full",
				Data:            getters.GetterKey[components.ObjectList[Teacher]]("teachers"),
				CreateUrl:       lago.GetterRoutePath("teachers.CreateRoute", nil),
				OnClick:         getters.GetterNavigateGetter(lago.GetterRoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))})),
				FilterComponent: lago.DynamicPage{Name: "teachers.TeacherFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "User.Name",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterForeignKey[p_users.User, uint, string](
									getters.GetterKey[uint]("$row.UserID"),
									"Name",
								),
							},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterKey[string]("$row.Code"),
							},
						},
					},
					{
						Label: "Email",
						Name:  "User.Email",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterForeignKey[p_users.User, uint, string](
									getters.GetterKey[uint]("$row.UserID"),
									"Email",
								),
							},
						},
					},
					{
						Label: "Qualifications",
						Name:  "Qualifications",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterDeref(getters.GetterKey[*string]("$row.Qualifications")),
							},
						},
					},
				},
			},
		},
	})
}

// --- Detail & Delete ---

func registerDetailPages() {
	lago.RegistryPage.Register("teachers.TeacherDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "teachers.TeacherDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Teacher]{
				Getter: getters.GetterKey[Teacher]("teacher"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{
							Key: "teachers.TeacherDetailContent",
						},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.GetterKey[string]("$in.User.Name"),
							},
							&components.FieldSubtitle{
								Getter: getters.GetterKey[string]("$in.Code"),
							},
							&components.LabelInline{
								Title:   "Email",
								Classes: "mt-4",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.GetterKey[string]("$in.User.Email"),
									},
								},
							},
							&components.LabelInline{
								Title: "Qualifications",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.GetterDeref(getters.GetterKey[*string]("$in.Qualifications")),
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("teachers.TeacherDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "teachers.TeacherDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:     "Confirm Deletion",
				Message:   "Are you sure you want to delete this teacher?",
				CancelUrl: lago.GetterRoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("teacher.ID"))}),
			},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("teachers.TeacherSelectionTable", &components.Modal{
		UID:   "teacher-selection-modal",
		Title: "Select Teacher",
		Children: []components.PageInterface{
			&components.DataTable[Teacher]{
				UID:  "teacher-selection-table",
				Data: getters.GetterKey[components.ObjectList[Teacher]]("teachers"),
				OnClick: getters.GetterSelect("TeacherID", getters.GetterKey[uint]("$row.ID"),
					getters.GetterFormat("%s (%s)",
						getters.GetterAny(getters.GetterForeignKey[p_users.User, uint, string](
							getters.GetterKey[uint]("$row.UserID"),
							"Name",
						)),
						getters.GetterAny(getters.GetterKey[string]("$row.Code")),
					),
				),
				FilterComponent: lago.DynamicPage{Name: "teachers.TeacherSelectionFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "User.Name",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterForeignKey[p_users.User, uint, string](
									getters.GetterKey[uint]("$row.UserID"),
									"Name",
								),
							},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterKey[string]("$row.Code"),
							},
						},
					},
					{
						Label: "Qualifications",
						Name:  "Qualifications",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterDeref(getters.GetterKey[*string]("$row.Qualifications")),
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("teachers.TeacherMultiSelectionTable", &components.Modal{
		UID:   "teacher-multi-selection-modal",
		Title: "Select Teachers",
		Children: []components.PageInterface{
			&components.DataTable[Teacher]{
				UID:  "teacher-multi-selection-table",
				Data: getters.GetterKey[components.ObjectList[Teacher]]("teachers"),
				OnClick: getters.GetterMultiSelect("Teachers",
					getters.GetterKey[uint]("$row.ID"),
					getters.GetterFormat("%s (%s)",
						getters.GetterAny(getters.GetterForeignKey[p_users.User, uint, string](
							getters.GetterKey[uint]("$row.UserID"),
							"Name",
						)),
						getters.GetterAny(getters.GetterKey[string]("$row.Code")),
					),
				),
				FilterComponent: lago.DynamicPage{Name: "teachers.TeacherMultiSelectionFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "User.Name",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterForeignKey[p_users.User, uint, string](
									getters.GetterKey[uint]("$row.UserID"),
									"Name",
								),
							},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterKey[string]("$row.Code"),
							},
						},
					},
					{
						Label: "Qualifications",
						Name:  "Qualifications",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterDeref(getters.GetterKey[*string]("$row.Qualifications")),
							},
						},
					},
				},
			},
		},
	})

}
