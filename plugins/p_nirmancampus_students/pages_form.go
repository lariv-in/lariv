package p_nirmancampus_students

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_users"
)

// studentFormUserPickURL opens the scoped user picker; on edit, allow_user_id keeps the linked user visible.
func studentFormUserPickURL(ctx context.Context) (string, error) {
	base, err := lago.RoutePath("students.UserPickRoute", nil)(ctx)
	if err != nil {
		return "", err
	}
	in, ok := ctx.Value(getters.ContextKeyIn).(map[string]any)
	if !ok {
		return base, nil
	}
	var uid uint
	switch v := in["UserID"].(type) {
	case uint:
		uid = v
	case uint8:
		uid = uint(v)
	case uint16:
		uid = uint(v)
	case uint32:
		uid = uint(v)
	case uint64:
		uid = uint(v)
	case float64:
		uid = uint(v)
	case int:
		if v > 0 {
			uid = uint(v)
		}
	case int64:
		if v > 0 {
			uid = uint(v)
		}
	case string:
		s := strings.TrimSpace(v)
		if s != "" {
			if n, err := strconv.ParseUint(s, 10, 64); err == nil {
				uid = uint(n)
			}
		}
	}
	if uid != 0 {
		return base + "?allow_user_id=" + strconv.FormatUint(uint64(uid), 10), nil
	}
	return base, nil
}
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
						Error: getters.Key[error]("$error.UserID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_users.User]{
								Label:       "User Account",
								Name:        "UserID",
								Required:    true,
								Getter:      getters.Association[p_users.User](getters.Key[uint]("$in.UserID")),
								Url:         studentFormUserPickURL,
								Display:     getters.Key[string]("$in.Name"),
								Placeholder: "Select a user...",
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.StudentNo"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Student Number",
								Name:     "StudentNo",
								Required: true,
								Getter:   getters.Key[string]("$in.StudentNo"),
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.DOB"),
				Children: []components.PageInterface{
					&components.InputDate{
						Label:  "Date of Birth",
						Name:   "DOB",
						Getter: getters.Deref(getters.Key[*time.Time]("$in.DOB")),
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
							&components.InputText{
								Label:  "Category",
								Name:   "Category",
								Getter: getters.Key[string]("$in.Category"),
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
				ActionURL: lago.RoutePath("students.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Student]{
						Attr: getters.FormBubbling(nil),

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
				ActionURL: lago.RoutePath("students.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[Student]{
						Getter: getters.Key[Student]("student"),
						Attr:   getters.FormBubbling(nil),

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
