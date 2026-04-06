package p_nirmancampus_courses

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

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
	createFormName := getters.Static("courses.CourseCreateForm")
	updateFormName := getters.Static("courses.CourseUpdateForm")
	deleteFormName := getters.Static("courses.CourseDeleteForm")

	lago.RegistryPage.Register("courses.CourseCreateForm", &components.Modal{
		Page: components.Page{
			Key:   "courses.CourseCreateModal",
			Roles: []string{"admin", "superuser"},
		},
		UID: "courses-create-modal",
		Children: []components.PageInterface{
			&components.FormComponent[Course]{
				Attr: getters.FormBubbling(createFormName),

				Title:    "Create Course",
				Subtitle: "Create a new course",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					courseFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ContainerRow{
						Classes: "flex justify-end gap-2 mt-2",
						Children: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save Course", Classes: "btn-primary"},
						},
					},
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
			&components.FormListenBoostedPost{
				Name:      updateFormName,
				ActionURL: lago.RoutePath("courses.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[Course]{
						Getter: getters.Key[Course]("course"),
						Attr:   getters.FormBubbling(updateFormName),

						Title:    "Edit Course",
						Subtitle: "Update course details",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							courseFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save Course"},
											&components.ButtonModalForm{
												Page:        components.Page{Roles: []string{"admin", "superuser"}},
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
			},
		},
	})
}
