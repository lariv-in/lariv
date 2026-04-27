package p_teachers

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("teachers.TeacherDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "teachers.TeacherDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[Teacher]{
				Getter: getters.Key[Teacher]("teacher"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "teachers.TeacherDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.LabelInline{Title: "Code", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.Code")}}},
							&components.LabelInline{Title: "Qualifications", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.Qualifications")}}},
							&components.LabelInline{Title: "Email", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.Email")}}},
							&components.LabelInline{Title: "Phone", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.Phone")}}},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("teachers.TeacherDeleteForm", &components.Modal{
		UID: "teacher-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Delete this teacher?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
