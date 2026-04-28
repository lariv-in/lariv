package p_semesters

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("semesters.SemesterDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "semesters.SemesterDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[Semester]{
				Getter: getters.Key[Semester]("semester"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "semesters.SemesterDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.LabelInline{Title: "Code", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.Code")}}},
							&components.LabelInline{Title: "Start", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.Start")}}},
							&components.LabelInline{Title: "End", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.End")}}},
							&components.LabelInline{Title: "Active", Children: []components.PageInterface{&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsActive")}}},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("semesters.SemesterDeleteForm", &components.Modal{
		UID: "semester-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm Deletion", Message: "Delete this semester?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
