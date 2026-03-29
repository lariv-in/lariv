package p_nirmancampus_courses

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
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
	lago.RegistryPage.Register("courses.CourseMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Courses"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to Home"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Courses"),
				Url:   lago.GetterRoutePath("courses.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Course: %s", getters.GetterAny(getters.GetterKey[string]("course.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Courses"),
			Url:   lago.GetterRoutePath("courses.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Course Detail"),
				Url:   lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("course.ID"))}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.GetterStatic("Edit Course"),
				Url:   lago.GetterRoutePath("courses.UpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("course.ID"))}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.GetterStatic("Delete Course"),
				Url:   lago.GetterRoutePath("courses.DeleteRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("course.ID"))}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("courses.CourseFilter", &components.FormComponent[Course]{
		Url:    lago.GetterRoutePath("courses.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$get.Name")},
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.GetterKey[string]("$get.Code")},
			&components.InputText{Label: "Type", Name: "CourseType", Getter: getters.GetterKey[string]("$get.CourseType")},
			&components.InputTernary{
				Label:      "Active",
				Name:       "IsActive",
				TrueLabel:  "Active Only",
				FalseLabel: "Inactive Only",
				NoneLabel:  "All",
				Getter:     getters.GetterKey[bool]("$get.IsActive"),
			},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply Filters"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("courses.CourseSelectionFilter", &components.FormComponent[Course]{
		Url:    lago.GetterRoutePath("courses.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$get.Name")},
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.GetterKey[string]("$get.Code")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("courses.CourseMultiSelectionFilter", &components.FormComponent[Course]{
		Url:    lago.GetterRoutePath("courses.MultiSelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$get.Name")},
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.GetterKey[string]("$get.Code")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})
}

// --- Form Fields & Forms ---

func courseFormFields() *components.ContainerColumn {
	return &components.ContainerColumn{
		Page: components.Page{
			Key: "courses.CourseFormFieldsBody",
		},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{Label: "Course Name", Name: "Name", Required: true, Getter: getters.GetterKey[string]("$in.Name")},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Code"),
						Children: []components.PageInterface{
							&components.InputText{Label: "Code", Name: "Code", Required: true, Getter: getters.GetterKey[string]("$in.Code")},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.CourseType"),
						Children: []components.PageInterface{
							&components.InputText{Label: "Type", Name: "CourseType", Getter: getters.GetterKey[string]("$in.CourseType")},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.IsActive"),
				Children: []components.PageInterface{
					&components.InputTernary{
						Label:      "Active",
						Name:       "IsActive",
						TrueLabel:  "Yes",
						FalseLabel: "No",
						NoneLabel:  "Not Set",
						Getter:     getters.GetterKey[bool]("$in.IsActive"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.Description"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:  "Description",
						Name:   "Description",
						Rows:   3,
						Getter: getters.GetterKey[string]("$in.Description"),
					},
				},
			},
		},
	}
}

func courseCreateUrlGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		role, err := getters.GetterKey[string]("$role")(ctx)
		if err != nil {
			return "", err
		}
		if role == "superuser" || role == "admin" {
			return lago.GetterRoutePath("courses.CreateRoute", nil)(ctx)
		}
		return "", fmt.Errorf("you do not have permission to do this action")
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("courses.CourseFormFields", courseFormFields())

	lago.RegistryPage.Register("courses.CourseCreateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Course]{
				Url:      lago.GetterRoutePath("courses.CreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create Course",
				Subtitle: "Create a new course",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					courseFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Course"},
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseUpdateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Course]{
				Getter:   getters.GetterKey[Course]("course"),
				Url:      lago.GetterRoutePath("courses.UpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}),
				Method:   http.MethodPost,
				Title:    "Edit Course",
				Subtitle: "Update course details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					courseFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Course"},
				},
			},
		},
	})
}

// --- Table ---

func registerTablePages() {
	lago.RegistryPage.Register("courses.CourseTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Course]{
				UID:             "course-table",
				Classes:         "w-full",
				Data:            getters.GetterKey[components.ObjectList[Course]]("courses"),
				CreateUrl:       courseCreateUrlGetter(),
				OnClick:         getters.GetterNavigateGetter(lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))})),
				FilterComponent: lago.DynamicPage{Name: "courses.CourseFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
					}},
					{Label: "Code", Name: "Code", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Code")},
					}},
					{Label: "Type", Name: "CourseType", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.CourseType")},
					}},
					{Label: "Active", Name: "IsActive", Children: []components.PageInterface{
						&components.FieldCheckbox{Getter: getters.GetterKey[bool]("$row.IsActive")},
					}},
				},
			},
		},
	})
}

// --- Detail & Delete ---

func registerDetailPages() {
	lago.RegistryPage.Register("courses.CourseDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Course]{
				Getter: getters.GetterKey[Course]("course"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page: components.Page{
							Key: "courses.CourseDetailContent",
						},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Code")},
							&components.LabelInline{
								Title:   "Type",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.CourseType")},
								},
							},
							&components.LabelInline{
								Title: "Active",
								Children: []components.PageInterface{
									&components.FieldCheckbox{Getter: getters.GetterKey[bool]("$in.IsActive")},
								},
							},
							&components.LabelInline{
								Title: "Description",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Description")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseDeleteForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:     "Confirm Deletion",
				Message:   "Are you sure you want to delete this course?",
				CancelUrl: lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("course.ID"))}),
			},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("courses.CourseSelectionTable", &components.Modal{
		UID:   "course-selection-modal",
		Title: "Select Course",
		Children: []components.PageInterface{
			&components.DataTable[Course]{
				UID:             "course-selection-table",
				Data:            getters.GetterKey[components.ObjectList[Course]]("courses"),
				OnClick:         getters.GetterSelect("CourseID", getters.GetterKey[uint]("$row.ID"), getters.GetterKey[string]("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "courses.CourseSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
					}},
					{Label: "Code", Name: "Code", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Code")},
					}},
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseMultiSelectionTable", &components.Modal{
		UID:   "course-multi-selection-modal",
		Title: "Select Courses",
		Children: []components.PageInterface{
			&components.DataTable[Course]{
				UID:             "course-multi-selection-table",
				Data:            getters.GetterKey[components.ObjectList[Course]]("courses"),
				OnClick:         getters.GetterMultiSelect("Courses", getters.GetterKey[uint]("$row.ID"), getters.GetterKey[string]("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "courses.CourseMultiSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
					}},
					{Label: "Code", Name: "Code", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Code")},
					}},
				},
			},
		},
	})
}
