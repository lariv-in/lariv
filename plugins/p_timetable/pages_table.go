package p_timetable

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("timetable.TimetableSlotTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "timetable.TimetableSlotMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[TimetableSlot]{
				Page: components.Page{Key: "timetable.TimetableSlotTableBody"}, UID: "timetable-slot-table", Classes: "w-full",
				Data:    getters.Key[components.ObjectList[TimetableSlot]]("timetable_slots"),
				Actions: []components.PageInterface{&components.TableButtonCreate{Link: lago.RoutePath("timetable.CreateRoute", nil)}},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("timetable.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Day", Name: "DayOfWeek", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.DayOfWeek")))}}},
					{Label: "Start min", Name: "StartMinute", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.StartMinute")))}}},
					{Label: "End min", Name: "EndMinute", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.EndMinute")))}}},
					{Label: "Label", Name: "Label", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Label")}}},
					{Label: "Course", Name: "Course", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Course.Code")}}},
				},
			},
		},
	})
}
