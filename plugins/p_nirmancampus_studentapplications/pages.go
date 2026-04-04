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
		Title: getters.Static("Student applications"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All applications"),
				Url:   lago.RoutePath("studentapplications.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("studentapplications.ApplicationDetailMenu", &components.SidebarMenu{
		Title: getters.Format(
			"Application: %s",
			getters.Any(getters.IfOrElse(
				getters.Key[string]("studentapplication.StudentName"),
				getters.IfOrElse(
					getters.Key[string]("$in.StudentName"),
					getters.Static("Application"),
				),
			)),
		),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all applications"),
			Url:   lago.RoutePath("studentapplications.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Application detail"),
				Url: lago.RoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.IfOrElse(
						getters.Key[uint]("studentapplication.ID"),
						getters.IfOrElse(
							getters.Key[uint]("$in.ID"),
							getters.ParseUint(getters.Key[string]("$path.id")),
						),
					)),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit application"),
				Url: lago.RoutePath("studentapplications.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.IfOrElse(
						getters.Key[uint]("studentapplication.ID"),
						getters.IfOrElse(
							getters.Key[uint]("$in.ID"),
							getters.ParseUint(getters.Key[string]("$path.id")),
						),
					)),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Delete application"),
				Url: lago.RoutePath("studentapplications.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.IfOrElse(
						getters.Key[uint]("studentapplication.ID"),
						getters.IfOrElse(
							getters.Key[uint]("$in.ID"),
							getters.ParseUint(getters.Key[string]("$path.id")),
						),
					)),
				}),
			},
		},
	})
}

func registerFilterPages() {
	lago.RegistryPage.Register("studentapplications.ApplicationFilter", &components.FormComponent[StudentApplication]{
		Url:    lago.RoutePath("studentapplications.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Email",
				Name:   "Email",
				Getter: getters.Key[string]("$get.Email"),
			},
			&components.InputText{
				Label:  "Student name",
				Name:   "StudentName",
				Getter: getters.Key[string]("$get.StudentName"),
			},
			&components.InputText{
				Label:  "Mobile",
				Name:   "Mobile",
				Getter: getters.Key[string]("$get.Mobile"),
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
		role, err := getters.Key[string]("$role")(ctx)
		if err != nil {
			return "", err
		}
		if role == "superuser" || role == "admin" || role == roleNameUnassigned {
			return lago.RoutePath("studentapplications.CreateRoute", nil)(ctx)
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
						Error: getters.Key[error]("$error.ProgramID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_nirmancampus_programs.Program]{
								Label:       "Program",
								Name:        "ProgramID",
								Required:    true,
								Getter:      getters.Association[p_nirmancampus_programs.Program](getters.Key[uint]("$in.ProgramID")),
								Url:         lago.RoutePath("programs.SelectRoute", nil),
								Display:     getters.Key[string]("$in.Name"),
								Placeholder: "Select a program...",
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.StudentName"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Student name",
								Name:     "StudentName",
								Required: true,
								Getter:   getters.Key[string]("$in.StudentName"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.DOB"),
						Children: []components.PageInterface{
							&components.InputDate{
								Label:    "Date of birth",
								Name:     "DOB",
								Required: false,
								Getter:   getters.Deref(getters.Key[*time.Time]("$in.DOB")),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.MotherName"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Mother name",
								Name:   "MotherName",
								Getter: getters.Key[string]("$in.MotherName"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.FatherName"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Father name",
								Name:   "FatherName",
								Getter: getters.Key[string]("$in.FatherName"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Category"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Category",
								Name:   "Category",
								Getter: getters.Key[string]("$in.Category"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Mobile"),
						Children: []components.PageInterface{
							&components.InputPhone{
								Label:  "Mobile",
								Name:   "Mobile",
								Getter: getters.Key[string]("$in.Mobile"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Email"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Email",
								Name:   "Email",
								Getter: getters.Key[string]("$in.Email"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Address"),
						Children: []components.PageInterface{
							&components.InputTextarea{
								Label:  "Address",
								Name:   "Address",
								Rows:   4,
								Getter: getters.Key[string]("$in.Address"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.PhotoID"),
						Children: []components.PageInterface{
							&p_filesystem.InputVNode{
								Label: "Photo",
								Name:  "PhotoID",
								VNode: func(ctx context.Context) (p_filesystem.VNode, error) {
									var zero p_filesystem.VNode
									if id, err := getters.Deref(getters.Key[*uint]("$in.PhotoID"))(ctx); err == nil && id != 0 {
										return getters.Association[p_filesystem.VNode](getters.Static(id))(ctx)
									}
									if id, err := getters.Key[uint]("$in.PhotoID")(ctx); err == nil && id != 0 {
										return getters.Association[p_filesystem.VNode](getters.Static(id))(ctx)
									}
									return zero, nil
								},
								AllowedFiletypes: []string{".jpg", ".jpeg", ".png", ".webp"},
								Path: getters.Format(
									"/studentapplications/%s-%d/",
									getters.Any(getters.IfOrElse(
										getters.Key[string]("$in.StudentName"),
										getters.Static("unknown"),
									)),
									getters.Any(getters.Key[int64]("$timestamp")),
								),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Documents"),
						Children: []components.PageInterface{
							&p_filesystem.InputMultiVNode{
								Label: "Documents",
								Name:  "Documents",
								VNode: func(ctx context.Context) ([]p_filesystem.VNode, error) {
									if nodes, err := getters.Key[[]p_filesystem.VNode]("$in.Documents")(ctx); err == nil && len(nodes) > 0 {
										return nodes, nil
									}
									return getters.AssociationList[p_filesystem.VNode](
										getters.AssociationIDs(getters.ContextKeyIn, "Documents"),
										"",
									)(ctx)
								},
								AllowedFiletypes: []string{".pdf", ".jpg", ".jpeg", ".png", ".webp"},
								Path: getters.Format(
									"/studentapplications/%s-%d/",
									getters.Any(getters.IfOrElse(
										getters.Key[string]("$in.StudentName"),
										getters.Static("unknown"),
									)),
									getters.Any(getters.Key[int64]("$timestamp")),
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
				Url:      lago.RoutePath("studentapplications.CreateRoute", nil),
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
				Getter: getters.Key[StudentApplication]("studentapplication"),
				Url: lago.RoutePath("studentapplications.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.IfOrElse(
						getters.Key[uint]("studentapplication.ID"),
						getters.IfOrElse(
							getters.Key[uint]("$in.ID"),
							getters.ParseUint(getters.Key[string]("$path.id")),
						),
					)),
				}),
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
				Page:    components.Page{Key: "studentapplications.ApplicationTableBody"},
				UID:     "student-application-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[StudentApplication]]("studentapplications"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "studentapplications.ApplicationFilter"}},
					&components.TableButtonCreate{Link: applicationCreateUrlGetter()},
				},
				OnClick: getters.NavigateGetter(lago.RoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{
						Label: "Email",
						Name:  "Email",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Email")},
						},
					},
					{
						Label: "Program",
						Name:  "Program.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Program.Name")},
						},
					},
					{
						Label: "Student",
						Name:  "StudentName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.StudentName")},
						},
					},
					{
						Label: "Mobile",
						Name:  "Mobile",
						Children: []components.PageInterface{
							&components.FieldPhone{Getter: getters.Key[string]("$row.Mobile")},
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
				Getter: getters.Key[StudentApplication]("studentapplication"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "studentapplications.ApplicationDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.Key[string]("$in.StudentName"),
							},
							&components.FieldSubtitle{
								Getter: getters.Key[string]("$in.Email"),
							},
							&components.LabelInline{
								Title: "Program",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Program.Name")},
								},
							},
							&components.LabelInline{
								Title: "Date of birth",
								Children: []components.PageInterface{
									&components.FieldDate{Getter: getters.Deref(getters.Key[*time.Time]("$in.DOB"))},
								},
							},
							&components.LabelInline{
								Title: "Mother name",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.MotherName")},
								},
							},
							&components.LabelInline{
								Title: "Father name",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.FatherName")},
								},
							},
							&components.LabelInline{
								Title: "Category",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Category")},
								},
							},
							&components.LabelInline{
								Title: "Mobile",
								Children: []components.PageInterface{
									&components.FieldPhone{Getter: getters.Key[string]("$in.Mobile")},
								},
							},
							&components.LabelInline{
								Title: "Email",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Email")},
								},
							},
							&components.LabelNewline{
								Title: "Address",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Address")},
								},
							},
							&components.LabelNewline{
								Title: "Photo",
								Children: []components.PageInterface{
									&p_filesystem.FieldPhoto{
										VNode:   getters.Association[p_filesystem.VNode](getters.Deref(getters.Key[*uint]("$in.PhotoID"))),
										Classes: "w-42 rounded",
									},
								},
							},
							&components.LabelNewline{
								Title: "Documents",
								Children: []components.PageInterface{
									&p_filesystem.FieldManyFile{
										VNode: getters.Key[[]p_filesystem.VNode]("$in.Documents"),
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
				Title:   "Confirm deletion",
				Message: "Are you sure you want to delete this application?",
				CancelUrl: lago.RoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.IfOrElse(
						getters.Key[uint]("studentapplication.ID"),
						getters.IfOrElse(
							getters.Key[uint]("$in.ID"),
							getters.ParseUint(getters.Key[string]("$path.id")),
						),
					)),
				}),
			},
		},
	})
}
