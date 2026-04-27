package p_timetable

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_semesters"
)

func timetableSlotFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "timetable.TimetableSlotFormFields"},
		Children: []components.PageInterface{
			&components.InputNumber[uint]{Label: "Day of week (0=Sun … 6=Sat)", Name: "DayOfWeek", Required: true, Getter: getters.Key[uint]("$in.DayOfWeek")},
			&components.InputNumber[uint]{Label: "Start minute", Name: "StartMinute", Required: true, Getter: getters.Key[uint]("$in.StartMinute")},
			&components.InputNumber[uint]{Label: "End minute", Name: "EndMinute", Required: true, Getter: getters.Key[uint]("$in.EndMinute")},
			&components.InputText{Label: "Label", Name: "Label", Getter: getters.Key[string]("$in.Label")},
			&components.InputForeignKey[p_courses.Course]{Label: "Course (optional)", Name: "CourseID", Required: false, Url: lago.RoutePath("courses.SelectRoute", nil), Display: getters.Key[string]("$in.Code"), Placeholder: "Select course...", Getter: getters.Association[p_courses.Course](getters.Deref(getters.Key[*uint]("$in.CourseID")))},
			&components.InputForeignKey[p_semesters.Semester]{Label: "Semester (optional)", Name: "SemesterID", Required: false, Url: lago.RoutePath("semesters.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "Semester…", Getter: getters.Association[p_semesters.Semester](getters.Deref(getters.Key[*uint]("$in.SemesterID")))},
		},
	}
}

func registerFormPages() {
	dn := getters.Static("timetable.TimetableSlotDeleteForm")
	lago.RegistryPage.Register("timetable.TimetableSlotCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "timetable.TimetableSlotMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("timetable.TimetableSlotCreateForm"),
				ActionURL: lago.RoutePath("timetable.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[TimetableSlot]{
						Attr:           getters.FormBubbling(getters.Static("timetable.TimetableSlotCreateForm")),
						Title:          "Create slot",
						ChildrenInput:  []components.PageInterface{timetableSlotFormFields()},
						ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save"}},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("timetable.TimetableSlotUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "timetable.TimetableSlotDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("timetable.TimetableSlotUpdateForm"),
				ActionURL: lago.RoutePath("timetable.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("timetable_slot.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[TimetableSlot]{
						Getter:        getters.Key[TimetableSlot]("timetable_slot"),
						Attr:          getters.FormBubbling(getters.Static("timetable.TimetableSlotUpdateForm")),
						Title:         "Edit slot",
						ChildrenInput: []components.PageInterface{timetableSlotFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label: "Delete", Icon: "trash", Name: dn,
										Url:         lago.RoutePath("timetable.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("timetable_slot.ID"))}),
										FormPostURL: lago.RoutePath("timetable.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("timetable_slot.ID"))}),
										ModalUID:    "timetable-slot-delete-modal", Classes: "btn-error",
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
