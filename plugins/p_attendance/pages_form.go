package p_attendance

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_programs"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"github.com/lariv-in/lago/plugins/p_sessions"
	"github.com/lariv-in/lago/plugins/p_students"
)

func attendanceMarkFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "attendance.AttendanceMarkFormFields"},
		Children: []components.PageInterface{
			&components.InputForeignKey[p_students.Student]{Label: "Student", Name: "StudentID", Required: true, Url: lago.RoutePath("students.SelectRoute", nil), Display: getters.Key[string]("$in.StudentNo"), Placeholder: "Select student...", Getter: getters.Association[p_students.Student](getters.Key[uint]("$in.StudentID"))},
			&components.InputForeignKey[p_sessions.ClassSession]{Label: "Session (optional)", Name: "SessionID", Required: false, Url: lago.RoutePath("sessions.SelectRoute", nil), Display: getters.Key[string]("$in.Title"), Placeholder: "Select session…", Getter: getters.Association[p_sessions.ClassSession](getters.Deref(getters.Key[*uint]("$in.SessionID")))},
			&components.InputForeignKey[p_courses.Course]{Label: "Course (optional)", Name: "CourseID", Required: false, Url: lago.RoutePath("courses.SelectRoute", nil), Display: getters.Key[string]("$in.Code"), Placeholder: "Select course…", Getter: getters.Association[p_courses.Course](getters.Deref(getters.Key[*uint]("$in.CourseID")))},
			&components.InputForeignKey[p_programs.Program]{Label: "Batch / program (optional)", Name: "BatchID", Required: false, Url: lago.RoutePath("programs.SelectRoute", nil), Display: getters.Key[string]("$in.Code"), Placeholder: "Select program…", Getter: getters.Association[p_programs.Program](getters.Deref(getters.Key[*uint]("$in.BatchID")))},
			&components.InputForeignKey[p_semesters.Semester]{Label: "Semester (optional)", Name: "SemesterID", Required: false, Url: lago.RoutePath("semesters.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "Semester…", Getter: getters.Association[p_semesters.Semester](getters.Deref(getters.Key[*uint]("$in.SemesterID")))},
			&components.InputCheckbox{Label: "Present", Name: "IsPresent", Getter: getters.Key[bool]("$in.IsPresent")},
			&components.InputTextarea{Label: "Notes", Name: "Notes", Rows: 3, Getter: getters.Key[string]("$in.Notes")},
			&components.ContainerError{Error: getters.Key[error]("$error.RecordedAt"), Children: []components.PageInterface{
				&components.InputDatetime{Label: "Recorded at", Name: "RecordedAt", Required: true, Getter: getters.Key[time.Time]("$in.RecordedAt")},
			}},
		},
	}
}

func registerFormPages() {
	dn := getters.Static("attendance.AttendanceMarkDeleteForm")
	lago.RegistryPage.Register("attendance.AttendanceMarkCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "attendance.AttendanceMarkMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("attendance.AttendanceMarkCreateForm"),
				ActionURL: lago.RoutePath("attendance.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[AttendanceMark]{
						Attr:           getters.FormBubbling(getters.Static("attendance.AttendanceMarkCreateForm")),
						Title:          "Create mark",
						ChildrenInput:  []components.PageInterface{attendanceMarkFormFields()},
						ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save"}},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("attendance.AttendanceMarkUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "attendance.AttendanceMarkDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("attendance.AttendanceMarkUpdateForm"),
				ActionURL: lago.RoutePath("attendance.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("attendance_mark.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[AttendanceMark]{
						Getter:        getters.Key[AttendanceMark]("attendance_mark"),
						Attr:          getters.FormBubbling(getters.Static("attendance.AttendanceMarkUpdateForm")),
						Title:         "Edit mark",
						ChildrenInput: []components.PageInterface{attendanceMarkFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label: "Delete", Icon: "trash", Name: dn,
										Url:         lago.RoutePath("attendance.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("attendance_mark.ID"))}),
										FormPostURL: lago.RoutePath("attendance.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("attendance_mark.ID"))}),
										ModalUID:    "attendance-mark-delete-modal", Classes: "btn-error",
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
