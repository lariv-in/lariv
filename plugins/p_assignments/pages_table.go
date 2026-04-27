package p_assignments

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("assignments.AssignmentTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "assignments.AssignmentMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[Assignment]{
				Page: components.Page{Key: "assignments.AssignmentTableBody"}, UID: "assignment-table", Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Assignment]]("assignments"),
				Actions: []components.PageInterface{&components.TableButtonCreate{Link: lago.RoutePath("assignments.CreateRoute", nil)}},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Course ID", Name: "CourseID", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.CourseID")))}}},
					{Label: "Title", Name: "Title", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Title")}}},
					{Label: "Due", Name: "DueAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Deref(getters.Key[*time.Time]("$row.DueAt"))}}},
				},
			},
		},
	})
}
