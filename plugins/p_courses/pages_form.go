package p_courses

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_programs"
	"github.com/lariv-in/lago/plugins/p_students"
)

func courseFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "courses.CourseFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{Error: getters.Key[error]("$error.Code"), Children: []components.PageInterface{
				&components.InputText{Label: "Code", Name: "Code", Required: true, Getter: getters.Key[string]("$in.Code")},
			}},
			&components.ContainerError{Error: getters.Key[error]("$error.Name"), Children: []components.PageInterface{
				&components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.Key[string]("$in.Name")},
			}},
			&components.InputCheckbox{Label: "Active", Name: "IsActive", Getter: getters.Key[bool]("$in.IsActive")},
			&components.InputTextarea{Label: "Description", Name: "Description", Rows: 3, Getter: getters.Key[string]("$in.Description")},
			&components.InputText{Label: "Subject", Name: "Subject", Getter: getters.Key[string]("$in.Subject")},
			&components.InputText{Label: "Course group", Name: "CourseGroup", Getter: getters.Key[string]("$in.CourseGroup")},
			&components.InputTextarea{Label: "Remarks", Name: "Remarks", Rows: 2, Getter: getters.Key[string]("$in.Remarks")},
			&components.InputText{Label: "Join code (UUID)", Name: "JoinCode", Getter: getters.Key[string]("$in.JoinCode")},
			&components.InputManyToMany[p_programs.Program]{
				Label: "Programs (batches)", Name: "Programs", Required: false,
				Getter: getters.Key[[]p_programs.Program]("$in.Programs"), Display: getters.Key[string]("$in.Code"),
				Url: lago.RoutePath("programs.MultiSelectRoute", nil), Placeholder: "Select programs…", Classes: "w-full",
			},
			&components.InputManyToMany[p_students.Student]{
				Label: "Students", Name: "Students", Required: false,
				Getter: getters.Key[[]p_students.Student]("$in.Students"), Display: getters.Key[string]("$in.StudentNo"),
				Url: lago.RoutePath("students.MultiSelectRoute", nil), Placeholder: "Select students…", Classes: "w-full",
			},
		},
	}
}

func registerFormPages() {
	deleteFormName := getters.Static("courses.CourseDeleteForm")

	lago.RegistryPage.Register("courses.CourseCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "courses.CourseMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("courses.CourseCreateForm"),
				ActionURL: lago.RoutePath("courses.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Course]{
						Attr:          getters.FormBubbling(getters.Static("courses.CourseCreateForm")),
						Title:         "Create Course",
						ChildrenInput: []components.PageInterface{courseFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save Course"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "courses.CourseDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("courses.CourseUpdateForm"),
				ActionURL: lago.RoutePath("courses.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("course.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Course]{
						Getter:        getters.Key[Course]("course"),
						Attr:          getters.FormBubbling(getters.Static("courses.CourseUpdateForm")),
						Title:         "Edit Course",
						ChildrenInput: []components.PageInterface{courseFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save Course"},
									&components.ButtonModalForm{
										Label:       "Delete",
										Icon:        "trash",
										Name:        deleteFormName,
										Url:         lago.RoutePath("courses.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))}),
										FormPostURL: lago.RoutePath("courses.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))}),
										ModalUID:    "course-delete-modal",
										Classes:     "btn-error",
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
