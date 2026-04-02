package p_nirmancampus_courses

import (
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
		Title: getters.Static("Courses"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Home"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Courses"),
				Url:   lago.RoutePath("courses.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Course: %s", getters.Any(getters.Key[string]("course.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Courses"),
			Url:   lago.RoutePath("courses.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Course Detail"),
				Url:   lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit Course"),
				Url:   lago.RoutePath("courses.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Delete Course"),
				Url:   lago.RoutePath("courses.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("courses.CourseFilter", &components.FormComponent[Course]{
		Url:    lago.RoutePath("courses.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.Key[string]("$get.Code")},
			&components.InputText{Label: "Type", Name: "CourseType", Getter: getters.Key[string]("$get.CourseType")},
			&components.InputTernary{
				Label:      "Active",
				Name:       "IsActive",
				TrueLabel:  "Active Only",
				FalseLabel: "Inactive Only",
				NoneLabel:  "All",
				Getter:     getters.Key[bool]("$get.IsActive"),
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
		Url:    lago.RoutePath("courses.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.Key[string]("$get.Code")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("courses.CourseMultiSelectionFilter", &components.FormComponent[Course]{
		Url:    lago.RoutePath("courses.MultiSelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Hidden: true, Name: "target_input", Getter: getters.Key[string]("$get.target_input")},
			&components.InputText{Hidden: true, Name: "pool_course_ids", Getter: getters.Key[string]("$get.pool_course_ids")},
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.Key[string]("$get.Code")},
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
						Error: getters.Key[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{Label: "Course Name", Name: "Name", Required: true, Getter: getters.Key[string]("$in.Name")},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Code"),
						Children: []components.PageInterface{
							&components.InputText{Label: "Code", Name: "Code", Required: true, Getter: getters.Key[string]("$in.Code")},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.CourseType"),
						Children: []components.PageInterface{
							&components.InputText{Label: "Type", Name: "CourseType", Getter: getters.Key[string]("$in.CourseType")},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.IsActive"),
				Children: []components.PageInterface{
					&components.InputTernary{
						Label:      "Active",
						Name:       "IsActive",
						TrueLabel:  "Yes",
						FalseLabel: "No",
						NoneLabel:  "Not Set",
						Getter:     getters.Key[bool]("$in.IsActive"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Description"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:  "Description",
						Name:   "Description",
						Rows:   3,
						Getter: getters.Key[string]("$in.Description"),
					},
				},
			},
		},
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
				Url:      lago.RoutePath("courses.CreateRoute", nil),
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
				Getter:   getters.Key[Course]("course"),
				Url:      lago.RoutePath("courses.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}),
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
				UID:     "course-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Course]]("courses"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "courses.CourseFilter"}},
					&components.TableButtonCreate{
						Page: components.Page{Roles: []string{"admin", "superuser"}},
						Link: lago.RoutePath("courses.CreateRoute", nil),
					},
				},
				OnClick: getters.NavigateGetter(lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
					{Label: "Code", Name: "Code", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Code")},
					}},
					{Label: "Type", Name: "CourseType", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.CourseType")},
					}},
					{Label: "Active", Name: "IsActive", Children: []components.PageInterface{
						&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsActive")},
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
				Getter: getters.Key[Course]("course"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page: components.Page{
							Key: "courses.CourseDetailContent",
						},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Code")},
							&components.LabelInline{
								Title:   "Type",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.CourseType")},
								},
							},
							&components.LabelInline{
								Title: "Active",
								Children: []components.PageInterface{
									&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsActive")},
								},
							},
							&components.LabelInline{
								Title: "Description",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Description")},
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
				CancelUrl: lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))}),
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
				UID:     "course-selection-table",
				Data:    getters.Key[components.ObjectList[Course]]("courses"),
				OnClick: getters.Select("CourseID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "courses.CourseSelectionFilter"}},
				},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
					{Label: "Code", Name: "Code", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Code")},
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
				UID:  "course-multi-selection-table",
				Data: getters.Key[components.ObjectList[Course]]("courses"),
				OnClick: getters.SelectMulti(
					getters.IfOrElse(
						getters.Key[string]("$get.target_input"),
						getters.Static("Courses"),
					),
					getters.Key[uint]("$row.ID"),
					getters.Key[string]("$row.Name"),
				),
				RowClass: getters.SelectMultiRowClass(
					getters.IfOrElse(
						getters.Key[string]("$get.target_input"),
						getters.Static("Courses"),
					),
					getters.Key[uint]("$row.ID"),
				),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "courses.CourseMultiSelectionFilter"}},
				},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
					{Label: "Code", Name: "Code", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Code")},
					}},
				},
			},
		},
	})
}
