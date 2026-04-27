package p_timetable

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("timetable.TimetableSlotDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "timetable.TimetableSlotDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[TimetableSlot]{
				Getter: getters.Key[TimetableSlot]("timetable_slot"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "timetable.TimetableSlotDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Label")},
							&components.LabelInline{Title: "Day of week", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.DayOfWeek")))},
							}},
							&components.LabelInline{Title: "Start minute", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.StartMinute")))},
							}},
							&components.LabelInline{Title: "End minute", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.EndMinute")))},
							}},
							&components.LabelInline{Title: "Course", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Course.Code")},
							}},
							&components.LabelInline{Title: "Semester", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Semester.Name")},
							}},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("timetable.TimetableSlotDeleteForm", &components.Modal{
		UID: "timetable-slot-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm deletion", Message: "Delete this slot?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
