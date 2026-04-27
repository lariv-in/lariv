package p_programs

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_students"
	"github.com/lariv-in/lago/plugins/p_teachers"
)

var (
	studentDetailLinkFromProgram = lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))})
	teacherDetailLinkFromProgram = lago.RoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))})
)

func registerDetailPages() {
	lago.RegistryPage.Register("programs.ProgramDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "programs.ProgramDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[Program]{
				Getter: getters.Key[Program]("program"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "programs.ProgramDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.LabelInline{Title: "Code", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.Code")}}},
							&components.FieldManyToMany[p_students.Student]{
								Label: "Students", Getter: getters.Key[[]p_students.Student]("$in.Students"),
								Display: getters.Key[string]("$in.StudentNo"), Link: studentDetailLinkFromProgram,
							},
							&components.FieldManyToMany[p_teachers.Teacher]{
								Label: "Teachers", Getter: getters.Key[[]p_teachers.Teacher]("$in.Teachers"),
								Display: getters.Key[string]("$in.Code"), Link: teacherDetailLinkFromProgram,
							},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("programs.ProgramDeleteForm", &components.Modal{
		UID: "program-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm Deletion", Message: "Delete this program?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
