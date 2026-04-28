package p_events

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("events.SchoolEventDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "events.SchoolEventDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[SchoolEvent]{
				Getter: getters.Key[SchoolEvent]("school_event"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "events.SchoolEventDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{Title: "Description", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Description")},
							}},
							&components.LabelInline{Title: "Starts", Children: []components.PageInterface{
								&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.StartsAt")},
							}},
							&components.LabelInline{Title: "Ends", Children: []components.PageInterface{
								&components.FieldDatetime{Getter: getters.Deref(getters.Key[*time.Time]("$in.EndsAt"))},
							}},
							&components.LabelInline{Title: "Universal", Children: []components.PageInterface{
								&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsUniversal")},
							}},
							&components.LabelInline{Title: "Semester", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Semester.Name")},
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
	lago.RegistryPage.Register("events.SchoolEventDeleteForm", &components.Modal{
		UID: "school-event-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm deletion", Message: "Delete this event?",
				Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
