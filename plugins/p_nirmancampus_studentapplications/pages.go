package p_nirmancampus_studentapplications

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
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
		Title: getters.GetterFormat("Application: %s", getters.GetterAny(getters.GetterKey[string]("studentapplication.StudentName"))),
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
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.GetterStatic("Edit application"),
				Url: lago.GetterRoutePath("studentapplications.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("studentapplication.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
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
				Label:  "Email",
				Name:   "Email",
				Getter: getters.GetterKey[string]("$get.Email"),
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

func applicationCreateUrlGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		role, err := getters.GetterKey[string]("$role")(ctx)
		if err != nil {
			return "", err
		}
		if role == "superuser" || role == "admin" || role == roleNameUnassigned {
			return lago.GetterRoutePath("studentapplications.CreateRoute", nil)(ctx)
		}
		return "", fmt.Errorf("you do not have permission to do this action")
	}
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
						Error: getters.GetterKey[error]("$error.ProgramID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_nirmancampus_programs.Program]{
								Label:       "Program",
								Name:        "ProgramID",
								Required:    true,
								Getter:      getters.GetterAssociation[p_nirmancampus_programs.Program](getters.GetterKey[uint]("$in.ProgramID")),
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
						Error: getters.GetterKey[error]("$error.DOB"),
						Children: []components.PageInterface{
							&components.InputDate{
								Label:    "Date of birth",
								Name:     "DOB",
								Required: false,
								Getter:   getters.GetterDeref(getters.GetterKey[*time.Time]("$in.DOB")),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.MotherName"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Mother name",
								Name:   "MotherName",
								Getter: getters.GetterKey[string]("$in.MotherName"),
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
							&components.InputPhone{
								Label:  "Mobile",
								Name:   "Mobile",
								Getter: getters.GetterKey[string]("$in.Mobile"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Email"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Email",
								Name:   "Email",
								Getter: getters.GetterKey[string]("$in.Email"),
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
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.PhotoID"),
						Children: []components.PageInterface{
							&p_filesystem.InputVNode{
								Label:            "Photo",
								Name:             "PhotoID",
								VNode:            getters.GetterAssociation[p_filesystem.VNode](getters.GetterDeref(getters.GetterKey[*uint]("$in.PhotoID"))),
								AllowedFiletypes: []string{".jpg", ".jpeg", ".png", ".webp"},
								Path: getters.GetterFormat(
									"/studentapplications/%s-%u/",
									getters.GetterAny(getters.GetterKey[string]("$in.StudentName")),
									getters.GetterAny(getters.GetterKey[int64]("$timestamp")),
								),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Documents"),
						Children: []components.PageInterface{
							&p_filesystem.InputMultiVNode{
								Label:            "Documents",
								Name:             "Documents",
								VNode:            getters.GetterKey[[]p_filesystem.VNode]("$in.Documents"),
								AllowedFiletypes: []string{".pdf", ".jpg", ".jpeg", ".png", ".webp"},
								Path: getters.GetterFormat(
									"/studentapplications/%s-%u/",
									getters.GetterAny(getters.GetterKey[string]("$in.StudentName")),
									getters.GetterAny(getters.GetterKey[int64]("$timestamp")),
								),
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
		Page: components.Page{Roles: []string{"admin", "superuser", roleNameUnassigned}},
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
		Page: components.Page{Roles: []string{"admin", "superuser"}},
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
				CreateUrl:       applicationCreateUrlGetter(),
				OnClick:         getters.GetterNavigateGetter(lago.GetterRoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))})),
				FilterComponent: lago.DynamicPage{Name: "studentapplications.ApplicationFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Email",
						Name:  "Email",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Email")},
						},
					},
					{
						Label: "Program",
						Name:  "Program.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Program.Name")},
						},
					},
					{
						Label: "Student",
						Name:  "StudentName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.StudentName")},
						},
					},
					{
						Label: "Mobile",
						Name:  "Mobile",
						Children: []components.PageInterface{
							&components.FieldPhone{Getter: getters.GetterKey[string]("$row.Mobile")},
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
								Getter: getters.GetterKey[string]("$in.StudentName"),
							},
							&components.FieldSubtitle{
								Getter: getters.GetterKey[string]("$in.Email"),
							},
							&components.LabelInline{
								Title: "Program",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Program.Name")},
								},
							},
							&components.LabelInline{
								Title: "Date of birth",
								Children: []components.PageInterface{
									&components.FieldDate{Getter: getters.GetterDeref(getters.GetterKey[*time.Time]("$in.DOB"))},
								},
							},
							&components.LabelInline{
								Title: "Mother name",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.MotherName")},
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
									&components.FieldPhone{Getter: getters.GetterKey[string]("$in.Mobile")},
								},
							},
							&components.LabelInline{
								Title: "Email",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Email")},
								},
							},
							&components.LabelInline{
								Title:   "Complete address",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.CompleteAddress")},
								},
							},
							&components.LabelInline{
								Title: "Photo",
								Children: []components.PageInterface{
									&p_filesystem.FieldPhoto{
										VNode:   getters.GetterAssociation[p_filesystem.VNode](getters.GetterDeref(getters.GetterKey[*uint]("$in.PhotoID"))),
										Classes: "w-48 rounded",
									},
								},
							},
							&components.LabelInline{
								Title: "Documents",
								Children: []components.PageInterface{
									&p_filesystem.FieldManyFile{
										VNode: getters.GetterKey[[]p_filesystem.VNode]("$in.Documents"),
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("studentapplications.ApplicationDeleteForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
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
