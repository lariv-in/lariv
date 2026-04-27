package p_assessments

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_students"
)

func gradeEntryFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "assessments.GradeEntryFormFields"},
		Children: []components.PageInterface{
			&components.InputForeignKey[p_students.Student]{Label: "Student", Name: "StudentID", Required: true, Url: lago.RoutePath("students.SelectRoute", nil), Display: getters.Key[string]("$in.StudentNo"), Placeholder: "Select student...", Getter: getters.Association[p_students.Student](getters.Key[uint]("$in.StudentID"))},
			&components.InputForeignKey[p_courses.Course]{Label: "Course", Name: "CourseID", Required: true, Url: lago.RoutePath("courses.SelectRoute", nil), Display: getters.Key[string]("$in.Code"), Placeholder: "Select course...", Getter: getters.Association[p_courses.Course](getters.Key[uint]("$in.CourseID"))},
			&components.InputText{Label: "Component", Name: "Component", Required: false, Getter: getters.Key[string]("$in.Component")},
			&components.InputNumber[float64]{Label: "Score", Name: "Score", Required: true, Getter: getters.Key[float64]("$in.Score")},
			&components.InputNumber[float64]{Label: "Max score", Name: "MaxScore", Required: true, Getter: getters.Key[float64]("$in.MaxScore")},
			&components.InputText{Label: "Status (PASS/FAIL)", Name: "Status", Getter: getters.Key[string]("$in.Status")},
			&components.InputTextarea{Label: "Remarks", Name: "Remarks", Rows: 3, Getter: getters.Key[string]("$in.Remarks")},
		},
	}
}

func registerFormPages() {
	dn := getters.Static("assessments.GradeEntryDeleteForm")
	lago.RegistryPage.Register("assessments.GradeEntryCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "assessments.GradeEntryMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("assessments.GradeEntryCreateForm"),
				ActionURL: lago.RoutePath("assessments.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[GradeEntry]{
						Attr:           getters.FormBubbling(getters.Static("assessments.GradeEntryCreateForm")),
						Title:          "Create grade entry",
						ChildrenInput:  []components.PageInterface{gradeEntryFormFields()},
						ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save"}},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("assessments.GradeEntryUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "assessments.GradeEntryDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("assessments.GradeEntryUpdateForm"),
				ActionURL: lago.RoutePath("assessments.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("grade_entry.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[GradeEntry]{
						Getter:        getters.Key[GradeEntry]("grade_entry"),
						Attr:          getters.FormBubbling(getters.Static("assessments.GradeEntryUpdateForm")),
						Title:         "Edit grade entry",
						ChildrenInput: []components.PageInterface{gradeEntryFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label: "Delete", Icon: "trash", Name: dn,
										Url:         lago.RoutePath("assessments.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("grade_entry.ID"))}),
										FormPostURL: lago.RoutePath("assessments.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("grade_entry.ID"))}),
										ModalUID:    "grade-entry-delete-modal", Classes: "btn-error",
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
