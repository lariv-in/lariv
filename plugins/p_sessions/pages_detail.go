package p_sessions

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("sessions.ClassSessionDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "sessions.ClassSessionDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[ClassSession]{
				Getter: getters.Key[ClassSession]("class_session"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "sessions.ClassSessionDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{Title: "Room", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Room")},
							}},
							&components.LabelInline{Title: "Start", Children: []components.PageInterface{
								&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.StartAt")},
							}},
							&components.LabelInline{Title: "End", Children: []components.PageInterface{
								&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.EndAt")},
							}},
							&components.LabelInline{Title: "Semester", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Semester.Name")},
							}},
							&components.LabelInline{Title: "Course", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Course.Code")},
							}},
							&components.LabelInline{Title: "Active", Children: []components.PageInterface{
								&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsActive")},
							}},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("sessions.ClassSessionDeleteForm", &components.Modal{
		UID: "class-session-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm deletion", Message: "Delete this session?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
