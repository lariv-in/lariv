package p_attendance

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("attendance.AttendanceMarkTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "attendance.AttendanceMarkMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[AttendanceMark]{
				Page: components.Page{Key: "attendance.AttendanceMarkTableBody"}, UID: "attendance-mark-table", Classes: "w-full",
				Data:    getters.Key[components.ObjectList[AttendanceMark]]("attendance_marks"),
				Actions: []components.PageInterface{&components.TableButtonCreate{Link: lago.RoutePath("attendance.CreateRoute", nil)}},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("attendance.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Student ID", Name: "StudentID", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.StudentID")))}}},
					{Label: "Recorded at", Name: "RecordedAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.RecordedAt")}}},
					{Label: "Present", Name: "IsPresent", Children: []components.PageInterface{&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsPresent")}}},
					{Label: "Notes", Name: "Notes", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Notes")}}},
				},
			},
		},
	})
}
