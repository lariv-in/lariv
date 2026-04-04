package p_nirmancampus_students

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_users"
)

func dobGetter() getters.Getter[time.Time] {
	return func(ctx context.Context) (time.Time, error) {
		dobPtr, err := getters.Key[*time.Time]("$in.DOB")(ctx)
		if err != nil || dobPtr == nil {
			return time.Time{}, nil
		}
		return *dobPtr, nil
	}
}

func dobDetailGetter() getters.Getter[time.Time] {
	return func(ctx context.Context) (time.Time, error) {
		dobPtr, err := getters.Key[*time.Time]("$in.DOB")(ctx)
		if err != nil || dobPtr == nil {
			return time.Time{}, nil
		}
		return *dobPtr, nil
	}
}

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
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
		Url:    lago.RoutePath("students.DefaultRoute", nil),
		Method: http.MethodGet,
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
				Label:  "Father's Name",
				Name:   "FathersName",
				Getter: getters.Key[string]("$get.FathersName"),
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
		Url:    lago.RoutePath("students.SelectRoute", nil),
		Method: http.MethodGet,
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
								Url:         lago.RoutePath("users.SelectRoute", nil),
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
						Getter: dobGetter(),
					},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.FathersName"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Father's Name",
								Name:   "FathersName",
								Getter: getters.Key[string]("$in.FathersName"),
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
						Rows:   3,
						Getter: getters.Key[string]("$in.Address"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Assets"),
				Children: []components.PageInterface{
					&components.InputManyToMany[p_filesystem.VNode]{
						Label:       "Assets",
						Name:        "Assets",
						Getter:      getters.Key[[]p_filesystem.VNode]("$in.Assets"),
						Url:         lago.RoutePath("filesystem.MultiSelectRoute", nil),
						Display:     getters.Key[string]("$in.Name"),
						Placeholder: "Select assets...",
					},
				},
			},
		},
	}
}

func studentCreateUrlGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		role, err := getters.Key[string]("$role")(ctx)
		if err != nil {
			return "", err
		}
		if role == "superuser" || role == "admin" {
			return lago.RoutePath("students.CreateRoute", nil)(ctx)
		}
		return "", fmt.Errorf("you do not have permission to do this action")
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
				Url:      lago.RoutePath("students.CreateRoute", nil),
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
				Url:      lago.RoutePath("students.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}),
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
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "students.StudentFilter"}, Page: components.Page{Roles: []string{"admin", "superuser"}}},
					&components.TableButtonCreate{Link: studentCreateUrlGetter(), Page: components.Page{Roles: []string{"admin", "superuser"}}},
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
						Label: "Father's Name",
						Name:  "FathersName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.FathersName")},
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
								Title: "Date of Birth",
								Children: []components.PageInterface{
									&components.FieldDate{
										Getter: dobDetailGetter(),
									},
								},
							},
							&components.LabelInline{
								Title: "Address",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Address")},
								},
							},
							&components.LabelInline{
								Title: "Category",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Category")},
								},
							},
							&components.LabelInline{
								Title: "Father's Name",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.FathersName")},
								},
							},
							&components.LabelInline{
								Title: "Assets",
								Children: []components.PageInterface{
									&components.FieldManyToMany[p_filesystem.VNode]{
										Getter:  getters.Key[[]p_filesystem.VNode]("$in.Assets"),
										Display: getters.Key[string]("$in.Name"),
										Link: lago.RoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
											"id": getters.Any(getters.Key[uint]("$in.ID")),
										}),
										Classes: "w-full",
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
						Label: "Father's Name",
						Name:  "FathersName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.FathersName")},
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
