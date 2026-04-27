package p_teachers

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func teacherFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "teachers.TeacherFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Code"),
				Children: []components.PageInterface{
					&components.InputText{Label: "Code", Name: "Code", Required: true, Getter: getters.Key[string]("$in.Code")},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.Key[string]("$in.Name")},
				},
			},
			&components.InputEmail{Label: "Email", Name: "Email", Getter: getters.Key[string]("$in.Email")},
			&components.InputPhone{Label: "Phone", Name: "Phone", Getter: getters.Key[string]("$in.Phone")},
			&components.InputTextarea{Label: "Qualifications", Name: "Qualifications", Rows: 4, Getter: getters.Key[string]("$in.Qualifications")},
		},
	}
}

func registerFormPages() {
	deleteFormName := getters.Static("teachers.TeacherDeleteForm")

	lago.RegistryPage.Register("teachers.TeacherCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "teachers.TeacherMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("teachers.TeacherCreateForm"),
				ActionURL: lago.RoutePath("teachers.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Teacher]{
						Attr:     getters.FormBubbling(getters.Static("teachers.TeacherCreateForm")),
						Title:    "Create Teacher",
						Subtitle: "Add a teacher record",
						ChildrenInput: []components.PageInterface{
							teacherFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save Teacher"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("teachers.TeacherUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "teachers.TeacherDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("teachers.TeacherUpdateForm"),
				ActionURL: lago.RoutePath("teachers.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("teacher.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Teacher]{
						Getter: getters.Key[Teacher]("teacher"),
						Attr:   getters.FormBubbling(getters.Static("teachers.TeacherUpdateForm")),
						Title:  "Edit Teacher",
						ChildrenInput: []components.PageInterface{
							teacherFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save Teacher"},
									&components.ButtonModalForm{
										Label:       "Delete",
										Icon:        "trash",
										Name:        deleteFormName,
										Url:         lago.RoutePath("teachers.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("teacher.ID"))}),
										FormPostURL: lago.RoutePath("teachers.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("teacher.ID"))}),
										ModalUID:    "teacher-delete-modal",
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
