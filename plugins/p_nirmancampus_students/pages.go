package p_nirmancampus_students

import (
	"context"
	"net/http"
	"strconv"
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
	}
	if uid != 0 {
		return base + "?allow_user_id=" + strconv.FormatUint(uint64(uid), 10), nil
	}
	return base, nil
}

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
	registerStudentUserPickPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("students.StudentMenu", &components.SidebarMenu{
		Title: getters.Static("Students"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Students"),
				Url:   lago.RoutePath("students.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("students.StudentDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Student: %s", getters.Any(getters.Key[string]("student.User.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Students"),
			Url:   lago.RoutePath("students.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Student Detail"),
				Url: lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("student.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit Student"),
				Url: lago.RoutePath("students.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("student.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Delete Student"),
				Url: lago.RoutePath("students.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("student.ID")),
				}),
			},
		},
	})
}

func registerFilterPages() {
	lago.RegistryPage.Register("students.StudentFilter", &components.FormComponent[Student]{
		OnSubmit: getters.FormSubmitGet(lago.RoutePath("students.DefaultRoute", nil)),
		Method:   http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Student Number",
				Name:   "StudentNo",
				Getter: getters.Key[string]("$get.StudentNo"),
			},
			&components.InputText{
				Label:  "Name",
				Name:   "User.Name",
				Getter: getters.Key[string]("$get.User.Name"),
			},
			&components.InputText{
				Label:  "Email",
				Name:   "User.Email",
				Getter: getters.Key[string]("$get.User.Email"),
			},
			&components.InputText{
				Label:  "Phone",
				Name:   "User.Phone",
				Getter: getters.Key[string]("$get.User.Phone"),
			},
			&components.InputText{
				Label:  "Mother's Name",
				Name:   "MotherName",
				Getter: getters.Key[string]("$get.MotherName"),
			},
			&components.InputText{
				Label:  "Father's Name",
				Name:   "FatherName",
				Getter: getters.Key[string]("$get.FatherName"),
			},
			&components.InputText{
				Label:  "Category",
				Name:   "Category",
				Getter: getters.Key[string]("$get.Category"),
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

	lago.RegistryPage.Register("students.StudentSelectionFilter", &components.FormComponent[Student]{
		OnSubmit: getters.FormSubmitGet(lago.RoutePath("students.SelectRoute", nil)),
		Method:   http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "User.Name",
				Getter: getters.Key[string]("$get.User.Name"),
			},
			&components.InputText{
				Label:  "Student No",
				Name:   "StudentNo",
				Getter: getters.Key[string]("$get.StudentNo"),
			},
			&components.InputText{
				Label:  "Phone",
				Name:   "User.Phone",
				Getter: getters.Key[string]("$get.User.Phone"),
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
			&components.FormComponent[Student]{
				OnSubmit: getters.FormSubmit(lago.RoutePath("students.CreateRoute", nil)),
				Method:   http.MethodPost,
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
	})

	lago.RegistryPage.Register("students.StudentUpdateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Student]{
				Getter:   getters.Key[Student]("student"),
				OnSubmit: getters.FormSubmit(lago.RoutePath("students.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))})),
				Method:   http.MethodPost,
				Title:    "Edit Student",
				Subtitle: "Update student details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					studentFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Student"},
				},
			},
		},
	})
}

func registerTablePages() {
	lago.RegistryPage.Register("students.StudentTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Student]{
				Page:    components.Page{Key: "students.StudentTableBody"},
				UID:     "student-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Student]]("students"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{
						Child: lago.DynamicPage{Name: "students.StudentFilter"},
						Page:  components.Page{Roles: []string{"admin", "superuser"}},
					},
					&components.TableButtonCreate{
						Link: lago.RoutePath("students.CreateRoute", nil),
						Page: components.Page{Roles: []string{"admin", "superuser"}},
					},
				},
				OnClick: getters.NavigateGetter(lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "User.Name",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.ForeignKey[p_users.User, uint, string](
									getters.Key[uint]("$row.UserID"),
									"Name",
								),
							},
						},
					},
					{
						Label: "Student No",
						Name:  "StudentNo",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Key[string]("$row.StudentNo"),
							},
						},
					},
					{
						Label: "Email",
						Name:  "User.Email",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.ForeignKey[p_users.User, uint, string](
									getters.Key[uint]("$row.UserID"),
									"Email",
								),
							},
						},
					},
					{
						Label: "Phone",
						Name:  "User.Phone",
						Children: []components.PageInterface{
							&components.FieldPhone{
								Getter: getters.ForeignKey[p_users.User, uint, string](
									getters.Key[uint]("$row.UserID"),
									"Phone",
								),
							},
						},
					},
					{
						Label: "Mother's Name",
						Name:  "MotherName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.MotherName")},
						},
					},
					{
						Label: "Father's Name",
						Name:  "FatherName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.FatherName")},
						},
					},
					{
						Label: "Category",
						Name:  "Category",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Category")},
						},
					},
					{
						Label: "Address",
						Name:  "Address",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Address")},
						},
					},
				},
			},
		},
	})
}

func registerDetailPages() {
	lago.RegistryPage.Register("students.StudentDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Student]{
				Getter: getters.Key[Student]("student"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "students.StudentDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.Key[string]("$in.User.Name"),
							},
							&components.FieldSubtitle{
								Getter: getters.Key[string]("$in.StudentNo"),
							},
							&components.LabelInline{
								Title: "Email",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.User.Email")},
								},
							},
							&components.LabelInline{
								Title: "Phone",
								Children: []components.PageInterface{
									&components.FieldPhone{Getter: getters.Key[string]("$in.User.Phone")},
								},
							},
							&components.LabelInline{
								Title: "Date of Birth",
								Children: []components.PageInterface{
									&components.FieldDate{
										Getter: getters.Deref(getters.Key[*time.Time]("$in.DOB")),
									},
								},
							},
							&components.LabelInline{
								Title: "Mother's Name",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.MotherName")},
								},
							},
							&components.LabelInline{
								Title: "Father's Name",
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

	lago.RegistryPage.Register("students.StudentDeleteForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:     "Confirm Deletion",
				Message:   "Are you sure you want to delete this student?",
				CancelUrl: lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student.ID"))}),
			},
		},
	})
}

func registerSelectionPages() {
	lago.RegistryPage.Register("students.StudentSelectionTable", &components.Modal{
		UID: "student-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Student]{
				Page:    components.Page{Key: "students.StudentSelectionTableBody"},
				UID:     "student-selection-table",
				Title:   "Select Student",
				Data:    getters.Key[components.ObjectList[Student]]("students"),
				OnClick: getters.Select("StudentID", getters.Key[uint]("$row.ID"), getters.ForeignKey[Student, uint, string](getters.Key[uint]("$row.ID"), "StudentNo")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "students.StudentSelectionFilter"}},
				},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "User.Name",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.ForeignKey[p_users.User, uint, string](
									getters.Key[uint]("$row.UserID"),
									"Name",
								),
							},
						},
					},
					{
						Label: "Student No",
						Name:  "StudentNo",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Key[string]("$row.StudentNo"),
							},
						},
					},
					{
						Label: "Phone",
						Name:  "User.Phone",
						Children: []components.PageInterface{
							&components.FieldPhone{
								Getter: getters.ForeignKey[p_users.User, uint, string](
									getters.Key[uint]("$row.UserID"),
									"Phone",
								),
							},
						},
					},
					{
						Label: "Mother's Name",
						Name:  "MotherName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.MotherName")},
						},
					},
					{
						Label: "Father's Name",
						Name:  "FatherName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.FatherName")},
						},
					},
					{
						Label: "Category",
						Name:  "Category",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Category")},
						},
					},
				},
			},
		},
	})
}

func registerStudentUserPickPages() {
	lago.RegistryPage.Register("students.UserPickFilter", &components.FormComponent[p_users.User]{
		OnSubmit: getters.FormSubmitGet(lago.RoutePath("students.UserPickRoute", nil)),
		Method:   http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Name:   "allow_user_id",
				Hidden: true,
				Getter: getters.Key[string]("$get.allow_user_id"),
			},
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
			&components.InputText{Label: "Email", Name: "Email", Getter: getters.Key[string]("$get.Email")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("students.UserPickTable", &components.Modal{
		UID: "student-user-pick-modal",
		Children: []components.PageInterface{
			&components.DataTable[p_users.User]{
				UID:     "student-user-pick-table",
				Title:   "Select User",
				Data:    getters.Key[components.ObjectList[p_users.User]]("users"),
				OnClick: getters.Select("UserID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "students.UserPickFilter"}},
					&components.ButtonModal{
						Url:     lago.RoutePath("users.CreateRoute", nil),
						Icon:    "plus",
						Classes: "btn-square btn-outline btn-sm",
					},
				},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
					{Label: "Email", Name: "Email", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Email")},
					}},
					{Label: "Phone", Name: "Phone", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Phone")},
					}},
				},
			},
		},
	})
}
