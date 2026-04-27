package p_courses

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_programs"
	"github.com/lariv-in/lago/plugins/p_students"
)

var (
	programDetailLink           = lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))})
	studentDetailLinkFromCourse = lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))})
)

func registerDetailPages() {
	lago.RegistryPage.Register("courses.CourseDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "courses.CourseDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[Course]{
				Getter: getters.Key[Course]("course"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "courses.CourseDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.LabelInline{
								Title: "Code",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Code")},
								},
							},
							&components.FieldManyToMany[p_programs.Program]{
								Label: "Programs (batches)", Getter: getters.Key[[]p_programs.Program]("$in.Programs"),
								Display: getters.Key[string]("$in.Code"),
								Link: programDetailLink,
							},
							&components.FieldManyToMany[p_students.Student]{
								Label: "Students", Getter: getters.Key[[]p_students.Student]("$in.Students"),
								Display: getters.Key[string]("$in.StudentNo"),
								Link:    studentDetailLinkFromCourse,
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseDeleteForm", &components.Modal{
		UID: "course-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Delete this course?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
