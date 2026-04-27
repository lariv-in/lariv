package p_students

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
)

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
						Page: components.Page{Key: "students.StudentDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.LabelInline{
								Title: "Student No",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.StudentNo")},
								},
							},
							&components.LabelInline{
								Title: "Email",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Email")},
								},
							},
							&components.LabelInline{
								Title: "Phone",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Phone")},
								},
							},
							&components.FieldManyToMany[p_filesystem.VNode]{
								Label: "Documents", Getter: getters.Key[[]p_filesystem.VNode]("$in.Documents"),
								Display: getters.Key[string]("$in.Name"),
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("students.StudentDeleteForm", &components.Modal{
		UID: "student-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Delete this student?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
