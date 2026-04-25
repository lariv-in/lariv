package p_nirmancampus_students

import (
	"context"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/registry"
)

func studentFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "students.StudentFormFieldsBody",
		},
		Children: []components.PageInterface{
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Name",
								Name:   "Name",
								Getter: getters.Key[string]("$in.Name"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.StudentNo"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Enrollment No / Control ID",
								Name:     "StudentNo",
								Required: true,
								Getter:   getters.Key[string]("$in.StudentNo"),
							},
						},
					},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.AadharCard"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Aadhar Card",
								Name:   "AadharCard",
								Getter: getters.Key[string]("$in.AadharCard"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.ABCId"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "ABC ID",
								Name:   "ABCId",
								Getter: getters.Key[string]("$in.ABCId"),
							},
						},
					},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
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
					&components.ContainerError{
						Error: getters.Key[error]("$error.Phone"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Phone",
								Name:   "Phone",
								Getter: getters.Key[string]("$in.Phone"),
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.DOB"),
				Children: []components.PageInterface{
					&components.InputDate{
						Label:    "Date of Birth",
						Name:     "DOB",
						Required: true,
						Getter:   getters.Deref(getters.Key[*time.Time]("$in.DOB")),
					},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.MotherName"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Mother's Name",
								Name:   "MotherName",
								Getter: getters.Key[string]("$in.MotherName"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.FatherName"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Father's Name",
								Name:   "FatherName",
								Getter: getters.Key[string]("$in.FatherName"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Category"),
						Children: []components.PageInterface{
							&components.InputSelect[string]{
								Label:    "Category",
								Name:     "Category",
								Required: false,
								Choices:  getters.Static(StudentCategoryChoices),
								Getter:   registry.PairFromGetter(getters.Key[string]("$in.Category"), StudentCategoryChoices),
							},
						},
					},
				},
			},
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
			&components.ContainerError{
				Error: getters.Key[error]("$error.Remarks"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:  "Remarks",
						Name:   "Remarks",
						Rows:   4,
						Getter: getters.Key[string]("$in.Remarks"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Handicapped"),
				Children: []components.PageInterface{
					&components.InputCheckbox{
						Label:  "Handicapped",
						Name:   "Handicapped",
						Getter: getters.Key[bool]("$in.Handicapped"),
					},
				},
			},
			components.ContainerRow{
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
									"/students/%s-%d/",
									getters.Any(getters.IfOrElse(
										getters.Key[string]("$in.StudentNo"),
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
									"/students/%s-%d/",
									getters.Any(getters.IfOrElse(
										getters.Key[string]("$in.StudentNo"),
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
	lago.RegistryPage.Register("students.StudentFormFields", studentFormFields())

	lago.RegistryPage.Register("students.StudentCreateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("students.StudentCreateForm"),
				ActionURL: lago.RoutePath("students.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Student]{
						Attr: getters.FormBubbling(getters.Static("students.StudentCreateForm")),

						Title:    "Create Student",
						Subtitle: "Create a new student",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							studentFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save Student"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("students.StudentUpdateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("students.StudentUpdateForm"),
				ActionURL: lago.RoutePath("students.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[Student]{
						Getter: getters.Key[Student]("student"),
						Attr:   getters.FormBubbling(getters.Static("students.StudentUpdateForm")),

						Title:    "Edit Student",
						Subtitle: "Update student details",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							studentFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save Student"},
											&components.ButtonModalForm{
												Page:        components.Page{Roles: []string{"admin", "superuser"}},
												Label:       "Delete",
												Icon:        "trash",
												Name:        getters.Static("students.StudentDeleteForm"),
												Url:         lago.RoutePath("students.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student.ID"))}),
												FormPostURL: lago.RoutePath("students.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student.ID"))}),
												ModalUID:    "student-delete-modal",
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
