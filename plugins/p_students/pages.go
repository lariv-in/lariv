package p_students

import (
	"context"
	"net/http"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_users"
)

// dobGetter returns a getter for the DOB field that handles the nullable *time.Time.
// It converts between *time.Time in the model and time.Time for the date input component.
func dobGetter() getters.Getter[time.Time] {
	return func(ctx context.Context) (time.Time, error) {
		dobPtr, err := getters.GetterKey[*time.Time]("$in.DOB")(ctx)
		if err != nil || dobPtr == nil {
			return time.Time{}, nil
		}
		return *dobPtr, nil
	}
}

// dobDetailGetter returns a getter for displaying DOB in the detail view.
// It reads from the student object (via $in) and handles the nullable time.
func dobDetailGetter() getters.Getter[time.Time] {
	return func(ctx context.Context) (time.Time, error) {
		dobPtr, err := getters.GetterKey[*time.Time]("$in.DOB")(ctx)
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

// --- Menus ---

func registerMenuPages() {
	lago.RegistryPage.Register("students.StudentMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Students"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Students"),
				Url:   lago.GetterRoutePath("students.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("students.StudentDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Student: %s", getters.GetterAny(getters.GetterKey[string]("student.User.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Students"),
			Url:   lago.GetterRoutePath("students.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Student Detail"),
				Url: lago.GetterRoutePath("students.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("student.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Student"),
				Url: lago.GetterRoutePath("students.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("student.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Student"),
				Url: lago.GetterRoutePath("students.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("student.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("students.StudentFilter", &components.FormComponent[Student]{
		Url:    lago.GetterRoutePath("students.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Student Number",
				Name:   "StudentNo",
				Getter: getters.GetterKey[string]("$get.StudentNo"),
			},
			&components.InputText{
				Label:  "Name",
				Name:   "User.Name",
				Getter: getters.GetterKey[string]("$get.User.Name"),
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
		Url:    lago.GetterRoutePath("students.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "User.Name",
				Getter: getters.GetterKey[string]("$get.User.Name"),
			},
			&components.InputText{
				Label:  "Student No",
				Name:   "StudentNo",
				Getter: getters.GetterKey[string]("$get.StudentNo"),
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
						Error: getters.GetterKey[error]("$error.StudentNo"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Student Number",
								Name:     "StudentNo",
								Required: true,
								Getter:   getters.GetterKey[string]("$in.StudentNo"),
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.DOB"),
				Children: []components.PageInterface{
					&components.InputDate{
						Label:  "Date of Birth",
						Name:   "DOB",
						Getter: dobGetter(),
					},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.Assets"),
				Children: []components.PageInterface{
					&components.InputManyToMany[p_filesystem.VNode]{
						Label:       "Assets",
						Name:        "Assets",
						Getter:      getters.GetterKey[[]p_filesystem.VNode]("$in.Assets"),
						Url:         lago.GetterRoutePath("filesystem.MultiSelectRoute", nil),
						Display:     getters.GetterKey[string]("$in.Name"),
						Placeholder: "Select assets...",
					},
				},
			},
		},
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("students.StudentFormFields", studentFormFields())

	lago.RegistryPage.Register("students.StudentCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Student]{
				Url:      lago.GetterRoutePath("students.CreateRoute", nil),
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
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Student]{
				Getter:   getters.GetterKey[Student]("student"),
				Url:      lago.GetterRoutePath("students.UpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}),
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

// --- Tables ---

func registerTablePages() {
	lago.RegistryPage.Register("students.StudentTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Student]{
				Page:            components.Page{Key: "students.StudentTableBody"},
				UID:             "student-table",
				Classes:         "w-full",
				Data:            getters.GetterKey[components.ObjectList[Student]]("students"),
				CreateUrl:       lago.GetterRoutePath("students.CreateRoute", nil),
				OnClick:         getters.GetterNavigateGetter(lago.GetterRoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))})),
				FilterComponent: lago.DynamicPage{Name: "students.StudentFilter"},
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
						Label: "Student No",
						Name:  "StudentNo",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterKey[string]("$row.StudentNo"),
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
				},
			},
		},
	})
}

// --- Detail & Delete ---

func registerDetailPages() {
	lago.RegistryPage.Register("students.StudentDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Student]{
				Getter: getters.GetterKey[Student]("student"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "students.StudentDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.GetterKey[string]("$in.User.Name"),
							},
							&components.FieldSubtitle{
								Getter: getters.GetterKey[string]("$in.StudentNo"),
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
								Title: "Assets",
								Children: []components.PageInterface{
									&components.FieldManyToMany[p_filesystem.VNode]{
										Getter:  getters.GetterKey[[]p_filesystem.VNode]("$in.Assets"),
										Display: getters.GetterKey[string]("$in.Name"),
										Link: lago.GetterRoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
											"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
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
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:     "Confirm Deletion",
				Message:   "Are you sure you want to delete this student?",
				CancelUrl: lago.GetterRoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("student.ID"))}),
			},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("students.StudentSelectionTable", &components.Modal{
		UID:   "student-selection-modal",
		Title: "Select Student",
		Children: []components.PageInterface{
			&components.DataTable[Student]{
				Page:            components.Page{Key: "students.StudentSelectionTableBody"},
				UID:             "student-selection-table",
				Data:            getters.GetterKey[components.ObjectList[Student]]("students"),
				OnClick:         getters.GetterSelect("StudentID", getters.GetterKey[uint]("$row.ID"), getters.GetterForeignKey[Student, uint, string](getters.GetterKey[uint]("$row.ID"), "StudentNo")),
				FilterComponent: lago.DynamicPage{Name: "students.StudentSelectionFilter"},
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
						Label: "Student No",
						Name:  "StudentNo",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterKey[string]("$row.StudentNo"),
							},
						},
					},
				},
			},
		},
	})
}
