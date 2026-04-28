package p_attendance

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("attendance.AttendanceMarkDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "attendance.AttendanceMarkDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[AttendanceMark]{
				Getter: getters.Key[AttendanceMark]("attendance_mark"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "attendance.AttendanceMarkDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Static("Attendance mark")},
							&components.LabelInline{Title: "Student", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Student.StudentNo")},
							}},
							&components.LabelInline{Title: "Course", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Course.Code")},
							}},
							&components.LabelInline{Title: "Program", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Program.Code")},
							}},
							&components.LabelInline{Title: "Semester", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Semester.Name")},
							}},
							&components.LabelInline{Title: "Session", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Session.Title")},
							}},
							&components.LabelInline{Title: "Recorded at", Children: []components.PageInterface{
								&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.RecordedAt")},
							}},
							&components.LabelInline{Title: "Present", Children: []components.PageInterface{
								&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsPresent")},
							}},
							&components.LabelInline{Title: "Notes", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Notes")},
							}},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("attendance.AttendanceMarkDeleteForm", &components.Modal{
		UID: "attendance-mark-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm deletion", Message: "Delete this mark?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
