package p_assignments

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

func registerDetailPages() {
	lago.RegistryPage.Register("assignments.AssignmentDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "assignments.AssignmentDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[Assignment]{
				Getter: getters.Key[Assignment]("assignment"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "assignments.AssignmentDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{Title: "Course", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Course.Code")},
							}},
							&components.LabelInline{Title: "Semester", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Semester.Name")},
							}},
							&components.LabelInline{Title: "Type", Children: []components.PageInterface{
								&components.FieldText{Getter: registry.PairValueFromKey(getters.Key[string]("$in.AssignmentType"), AssignmentTypeChoices)},
							}},
							&components.LabelInline{Title: "Total marks", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[int]("$in.TotalMarks")))},
							}},
							&components.LabelInline{Title: "Release at", Children: []components.PageInterface{
								&components.FieldDatetime{Getter: getters.Deref(getters.Key[*time.Time]("$in.ReleaseAt"))},
							}},
							&components.LabelInline{Title: "Due at", Children: []components.PageInterface{
								&components.FieldDatetime{Getter: getters.Deref(getters.Key[*time.Time]("$in.DueAt"))},
							}},
							&components.LabelInline{Title: "Description", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Description")},
							}},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("assignments.AssignmentDeleteForm", &components.Modal{
		UID: "assignment-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm deletion", Message: "Delete this assignment?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
