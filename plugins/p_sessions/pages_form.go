package p_sessions

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_semesters"
)

func classSessionFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "sessions.ClassSessionFormFields"},
		Children: []components.PageInterface{
			&components.InputText{Label: "Title", Name: "Title", Required: true, Getter: getters.Key[string]("$in.Title")},
			&components.InputText{Label: "Room", Name: "Room", Getter: getters.Key[string]("$in.Room")},
			&components.InputDatetime{Label: "Start at", Name: "StartAt", Required: true, Getter: getters.Key[time.Time]("$in.StartAt")},
			&components.InputDatetime{Label: "End at", Name: "EndAt", Required: true, Getter: getters.Key[time.Time]("$in.EndAt")},
			&components.InputCheckbox{Label: "Active", Name: "IsActive", Getter: getters.Key[bool]("$in.IsActive")},
			&components.InputForeignKey[p_semesters.Semester]{Label: "Semester (optional)", Name: "SemesterID", Required: false, Url: lago.RoutePath("semesters.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "Semester…", Getter: getters.Association[p_semesters.Semester](getters.Deref(getters.Key[*uint]("$in.SemesterID")))},
			&components.InputForeignKey[p_courses.Course]{Label: "Course (optional)", Name: "CourseID", Required: false, Url: lago.RoutePath("courses.SelectRoute", nil), Display: getters.Key[string]("$in.Code"), Placeholder: "Course…", Getter: getters.Association[p_courses.Course](getters.Deref(getters.Key[*uint]("$in.CourseID")))},
		},
	}
}

func registerFormPages() {
	dn := getters.Static("sessions.ClassSessionDeleteForm")
	lago.RegistryPage.Register("sessions.ClassSessionCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "sessions.ClassSessionMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("sessions.ClassSessionCreateForm"),
				ActionURL: lago.RoutePath("sessions.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[ClassSession]{
						Attr:           getters.FormBubbling(getters.Static("sessions.ClassSessionCreateForm")),
						Title:          "Create session",
						ChildrenInput:  []components.PageInterface{classSessionFormFields()},
						ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save"}},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("sessions.ClassSessionUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "sessions.ClassSessionDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("sessions.ClassSessionUpdateForm"),
				ActionURL: lago.RoutePath("sessions.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("class_session.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[ClassSession]{
						Getter:        getters.Key[ClassSession]("class_session"),
						Attr:          getters.FormBubbling(getters.Static("sessions.ClassSessionUpdateForm")),
						Title:         "Edit session",
						ChildrenInput: []components.PageInterface{classSessionFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label: "Delete", Icon: "trash", Name: dn,
										Url:         lago.RoutePath("sessions.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("class_session.ID"))}),
										FormPostURL: lago.RoutePath("sessions.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("class_session.ID"))}),
										ModalUID:    "class-session-delete-modal", Classes: "btn-error",
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
