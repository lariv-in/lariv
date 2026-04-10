package p_nirmancampus_studentapplications

import (
	"context"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
)

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
			&components.FormListenBoostedPost{
				Name:      getters.Static("studentapplications.ApplicationCreateForm"),
				ActionURL: lago.RoutePath("studentapplications.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[StudentApplication]{
						Attr: getters.FormBubbling(getters.Static("studentapplications.ApplicationCreateForm")),

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
			},
		},
	})

	lago.RegistryPage.Register("studentapplications.ApplicationUpdateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "studentapplications.ApplicationDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("studentapplications.ApplicationUpdateForm"),
				ActionURL: lago.RoutePath("studentapplications.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.IfOrElse(
						getters.Key[uint]("studentapplication.ID"),
						getters.ParseUint(getters.Key[string]("$path.id")),
					)),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[StudentApplication]{
						Getter: getters.Key[StudentApplication]("studentapplication"),
						Attr:   getters.FormBubbling(getters.Static("studentapplications.ApplicationUpdateForm")),

						Title:    "Edit application",
						Subtitle: "Update application details",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							applicationFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save application"},
											&components.ButtonModalForm{
												Page:        components.Page{Roles: []string{"admin", "superuser"}},
												Label:       "Delete",
												Icon:        "trash",
												Name:        getters.Static("studentapplications.ApplicationDeleteForm"),
												Url:         lago.RoutePath("studentapplications.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("studentapplication.ID"))}),
												FormPostURL: lago.RoutePath("studentapplications.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("studentapplication.ID"))}),
												ModalUID:    "studentapplication-delete-modal",
												Classes:     "btn-error",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
}
