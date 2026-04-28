package p_assessments

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("assessments.GradeEntryTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "assessments.GradeEntryMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[GradeEntry]{
				Page: components.Page{Key: "assessments.GradeEntryTableBody"}, UID: "grade-entry-table", Classes: "w-full",
				Data:    getters.Key[components.ObjectList[GradeEntry]]("grade_entries"),
				Actions: []components.PageInterface{&components.TableButtonCreate{Link: lago.RoutePath("assessments.CreateRoute", nil)}},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("assessments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Student", Name: "StudentID", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.StudentID")))}}},
					{Label: "Course", Name: "CourseID", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.CourseID")))}}},
					{Label: "Component", Name: "Component", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Component")}}},
					{Label: "Score", Name: "Score", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%.2f", getters.Any(getters.Key[float64]("$row.Score")))}}},
					{Label: "Max", Name: "MaxScore", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%.2f", getters.Any(getters.Key[float64]("$row.MaxScore")))}}},
				},
			},
		},
	})
}
