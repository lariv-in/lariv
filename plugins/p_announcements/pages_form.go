package p_announcements

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
)

func announcementFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "announcements.AnnouncementFormFields"},
		Children: []components.PageInterface{
			&components.InputText{Label: "Title", Name: "Title", Required: true, Getter: getters.Key[string]("$in.Title")},
			&components.InputTextarea{Label: "Description", Name: "Description", Rows: 5, Getter: getters.Key[string]("$in.Description")},
			&components.InputCheckbox{Label: "Universal", Name: "IsUniversal", Getter: getters.Key[bool]("$in.IsUniversal")},
			&components.InputForeignKey[p_semesters.Semester]{Label: "Semester (optional)", Name: "SemesterID", Required: false, Url: lago.RoutePath("semesters.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "Semester…", Getter: getters.Association[p_semesters.Semester](getters.Deref(getters.Key[*uint]("$in.SemesterID")))},
			&components.InputForeignKey[p_users.User]{Label: "Created by (optional)", Name: "CreatedByID", Required: false, Url: lago.RoutePath("users.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "User…", Getter: getters.Association[p_users.User](getters.Deref(getters.Key[*uint]("$in.CreatedByID")))},
			&components.InputForeignKey[p_users.User]{Label: "Signed by (optional)", Name: "SignedByID", Required: false, Url: lago.RoutePath("users.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "User…", Getter: getters.Association[p_users.User](getters.Deref(getters.Key[*uint]("$in.SignedByID")))},
			&components.ContainerError{Error: getters.Key[error]("$error.ReleaseAt"), Children: []components.PageInterface{
				&components.InputDatetime{Label: "Release at", Name: "ReleaseAt", Required: true, Getter: getters.Key[time.Time]("$in.ReleaseAt")},
			}},
			&components.InputDatetime{Label: "Expiry at (optional)", Name: "ExpiryAt", Required: false, Getter: getters.Deref(getters.Key[*time.Time]("$in.ExpiryAt"))},
			&components.InputSelect[string]{Label: "Priority", Name: "Priority", Required: false, Choices: getters.Static(AnnouncementPriorityChoices), Getter: registry.PairFromGetter(getters.Key[string]("$in.Priority"), AnnouncementPriorityChoices)},
		},
	}
}

func registerFormPages() {
	dn := getters.Static("announcements.AnnouncementDeleteForm")
	lago.RegistryPage.Register("announcements.AnnouncementCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "announcements.AnnouncementMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("announcements.AnnouncementCreateForm"),
				ActionURL: lago.RoutePath("announcements.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Announcement]{
						Attr:           getters.FormBubbling(getters.Static("announcements.AnnouncementCreateForm")),
						Title:          "Create announcement",
						ChildrenInput:  []components.PageInterface{announcementFormFields()},
						ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save"}},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("announcements.AnnouncementUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "announcements.AnnouncementDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("announcements.AnnouncementUpdateForm"),
				ActionURL: lago.RoutePath("announcements.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("announcement.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[Announcement]{
						Getter:        getters.Key[Announcement]("announcement"),
						Attr:          getters.FormBubbling(getters.Static("announcements.AnnouncementUpdateForm")),
						Title:         "Edit announcement",
						ChildrenInput: []components.PageInterface{announcementFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label: "Delete", Icon: "trash", Name: dn,
										Url:         lago.RoutePath("announcements.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("announcement.ID"))}),
										FormPostURL: lago.RoutePath("announcements.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("announcement.ID"))}),
										ModalUID:    "announcement-delete-modal", Classes: "btn-error",
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
