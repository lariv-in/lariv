package p_allocation

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("allocation.CourseTeacherAssignmentTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "allocation.CourseTeacherAssignmentMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[CourseTeacherAssignment]{
				Page: components.Page{Key: "allocation.CourseTeacherAssignmentTableBody"}, UID: "allocation-table", Classes: "w-full",
				Data:    getters.Key[components.ObjectList[CourseTeacherAssignment]]("allocations"),
				Actions: []components.PageInterface{&components.TableButtonCreate{Link: lago.RoutePath("allocation.CreateRoute", nil)}},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("allocation.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Teacher ID", Name: "TeacherID", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.TeacherID")))}}},
					{Label: "Course ID", Name: "CourseID", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.CourseID")))}}},
					{Label: "Role", Name: "Role", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Role")}}},
				},
			},
		},
	})
}
