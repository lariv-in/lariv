package p_nirmancampus_announcements

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("announcements.AnnouncementDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "announcements.AnnouncementDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Announcement]{
				Getter: getters.Key[Announcement]("announcement"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page: components.Page{Key: "announcements.AnnouncementDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Page:   components.Page{Key: "announcements.AnnouncementDetailTitle"},
								Getter: getters.Key[string]("$in.Title"),
							},
							&components.LabelNewline{
								Title: "Description",
								Children: []components.PageInterface{
									&components.FieldMarkdown{Getter: getters.Key[string]("$in.Description")},
								},
							},
							&components.LabelInline{
								Title: "URL",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.URL")},
								},
							},
							&components.LabelInline{
								Title: "Release At",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.ReleaseAt")},
								},
							},
							&components.LabelInline{
								Title: "Expiry At",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.Deref(getters.Key[*time.Time]("$in.ExpiryAt"))},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("announcements.AnnouncementDeleteForm", &components.Modal{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		UID:  "announcement-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this announcement?",
				Attr:    getters.FormBubbling(nil),
			},
		},
	})
}
