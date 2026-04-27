package p_assessments

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("assessments.GradeEntryDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "assessments.GradeEntryDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[GradeEntry]{
				Getter: getters.Key[GradeEntry]("grade_entry"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "assessments.GradeEntryDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Component")},
							&components.LabelInline{Title: "Student ID", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.StudentID")))},
							}},
							&components.LabelInline{Title: "Course ID", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.CourseID")))},
							}},
							&components.LabelInline{Title: "Score", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%.2f", getters.Any(getters.Key[float64]("$in.Score")))},
							}},
							&components.LabelInline{Title: "Max score", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%.2f", getters.Any(getters.Key[float64]("$in.MaxScore")))},
							}},
							&components.LabelInline{Title: "Status", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Status")},
							}},
							&components.LabelInline{Title: "Remarks", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Remarks")},
							}},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("assessments.GradeEntryDeleteForm", &components.Modal{UID: "grade-entry-delete-modal", Children: []components.PageInterface{
		&components.DeleteConfirmation{Title: "Confirm deletion", Message: "Delete this grade entry?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
	}})
}
