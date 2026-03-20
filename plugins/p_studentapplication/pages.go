package p_studentapplication

import (
	"net/http"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_programs"
)

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("studentapplications.ApplicationMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Student applications"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All applications"),
				Url:   lago.GetterRoutePath("studentapplications.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("studentapplications.ApplicationDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Application: %s", getters.GetterAny(getters.GetterKey[string]("studentapplication.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all applications"),
			Url:   lago.GetterRoutePath("studentapplications.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Application detail"),
				Url: lago.GetterRoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("studentapplication.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit application"),
				Url: lago.GetterRoutePath("studentapplications.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("studentapplication.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete application"),
				Url: lago.GetterRoutePath("studentapplications.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("studentapplication.ID")),
				}),
			},
		},
	})
}

func registerFilterPages() {
	lago.RegistryPage.Register("studentapplications.ApplicationFilter", &components.FormComponent[StudentApplication]{
		Url:    lago.GetterRoutePath("studentapplications.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.GetterKey[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Student name",
				Name:   "StudentName",
				Getter: getters.GetterKey[string]("$get.StudentName"),
			},
			&components.InputText{
				Label:  "Mobile",
				Name:   "Mobile",
				Getter: getters.GetterKey[string]("$get.Mobile"),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply filters"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})
}

func applicationFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "studentapplications.ApplicationFormFieldsBody",
		},
		Children: []components.PageInterface{
			&components.ContainerRow{
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
						Error: getters.GetterKey[error]("$error.ProgramID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_programs.Program]{
								Label:       "Program",
								Name:        "ProgramID",
								Required:    true,
								Getter:      getters.GetterAssociation[p_programs.Program](getters.GetterKey[uint]("$in.ProgramID")),
								Url:         lago.GetterRoutePath("programs.SelectRoute", nil),
								Display:     getters.GetterKey[string]("$in.Name"),
								Placeholder: "Select a program...",
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.StudentName"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Student name",
								Name:     "StudentName",
								Required: true,
								Getter:   getters.GetterKey[string]("$in.StudentName"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.FatherName"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Father name",
								Name:   "FatherName",
								Getter: getters.GetterKey[string]("$in.FatherName"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Category"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Category",
								Name:   "Category",
								Getter: getters.GetterKey[string]("$in.Category"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Mobile"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Mobile",
								Name:   "Mobile",
								Getter: getters.GetterKey[string]("$in.Mobile"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.CompleteAddress"),
						Children: []components.PageInterface{
							&components.InputTextarea{
								Label:  "Complete address",
								Name:   "CompleteAddress",
								Rows:   4,
								Getter: getters.GetterKey[string]("$in.CompleteAddress"),
							},
						},
					},
				},
			},
		},
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("studentapplications.ApplicationFormFields", applicationFormFields())

	lago.RegistryPage.Register("studentapplications.ApplicationCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "studentapplications.ApplicationMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[StudentApplication]{
				Url:      lago.GetterRoutePath("studentapplications.CreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create application",
				Subtitle: "Record a new student application",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					applicationFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save application"},
				},
			},
		},
	})

	lago.RegistryPage.Register("studentapplications.ApplicationUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "studentapplications.ApplicationDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[StudentApplication]{
				Getter:   getters.GetterKey[StudentApplication]("studentapplication"),
				Url:      lago.GetterRoutePath("studentapplications.UpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}),
				Method:   http.MethodPost,
				Title:    "Edit application",
				Subtitle: "Update application details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					applicationFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save application"},
				},
			},
		},
	})
}

func registerTablePages() {
	lago.RegistryPage.Register("studentapplications.ApplicationTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "studentapplications.ApplicationMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[StudentApplication]{
				Page:            components.Page{Key: "studentapplications.ApplicationTableBody"},
				UID:             "student-application-table",
				Classes:         "w-full",
				Data:            getters.GetterKey[components.ObjectList[StudentApplication]]("studentapplications"),
				CreateUrl:       lago.GetterRoutePath("studentapplications.CreateRoute", nil),
				OnClick:         getters.GetterNavigateGetter(lago.GetterRoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))})),
				FilterComponent: lago.DynamicPage{Name: "studentapplications.ApplicationFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Key:   "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
						},
					},
					{
						Label: "Program",
						Key:   "Program.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Program.Name")},
						},
					},
					{
						Label: "Student",
						Key:   "StudentName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.StudentName")},
						},
					},
					{
						Label: "Mobile",
						Key:   "Mobile",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Mobile")},
						},
					},
				},
			},
		},
	})
}

func registerDetailPages() {
	lago.RegistryPage.Register("studentapplications.ApplicationDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "studentapplications.ApplicationDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[StudentApplication]{
				Getter: getters.GetterKey[StudentApplication]("studentapplication"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "studentapplications.ApplicationDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.GetterKey[string]("$in.Name"),
							},
							&components.FieldSubtitle{
								Getter: getters.GetterKey[string]("$in.StudentName"),
							},
							&components.LabelInline{
								Title: "Program",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Program.Name")},
								},
							},
							&components.LabelInline{
								Title: "Father name",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.FatherName")},
								},
							},
							&components.LabelInline{
								Title: "Category",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Category")},
								},
							},
							&components.LabelInline{
								Title: "Mobile",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Mobile")},
								},
							},
							&components.LabelInline{
								Title:   "Complete address",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.CompleteAddress")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("studentapplications.ApplicationDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "studentapplications.ApplicationDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:     "Confirm deletion",
				Message:   "Are you sure you want to delete this application?",
				CancelUrl: lago.GetterRoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("studentapplication.ID"))}),
			},
		},
	})
}
