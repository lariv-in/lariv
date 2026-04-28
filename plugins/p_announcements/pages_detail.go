package p_announcements

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

func registerDetailPages() {
	lago.RegistryPage.Register("announcements.AnnouncementDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "announcements.AnnouncementDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[Announcement]{
				Getter: getters.Key[Announcement]("announcement"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "announcements.AnnouncementDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{Title: "Description", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Description")},
							}},
							&components.LabelInline{Title: "Universal", Children: []components.PageInterface{
								&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsUniversal")},
							}},
							&components.LabelInline{Title: "Semester", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Semester.Name")},
							}},
							&components.LabelInline{Title: "Created by", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.CreatedBy.Name")},
							}},
							&components.LabelInline{Title: "Signed by", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.SignedBy.Name")},
							}},
							&components.LabelInline{Title: "Release at", Children: []components.PageInterface{
								&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.ReleaseAt")},
							}},
							&components.LabelInline{Title: "Expiry at", Children: []components.PageInterface{
								&components.FieldDatetime{Getter: getters.Deref(getters.Key[*time.Time]("$in.ExpiryAt"))},
							}},
							&components.LabelInline{Title: "Priority", Children: []components.PageInterface{
								&components.FieldText{Getter: registry.PairValueFromKey(getters.Key[string]("$in.Priority"), AnnouncementPriorityChoices)},
							}},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("announcements.AnnouncementDeleteForm", &components.Modal{
		UID: "announcement-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm deletion", Message: "Delete this announcement?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
