package p_nirmancampus_announcements

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func announcementFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "announcements.AnnouncementFormFieldsBody"},
		Children: []components.PageInterface{
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Title"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Title",
								Name:     "Title",
								Required: true,
								Getter:   getters.Key[string]("$in.Title"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Description"),
						Children: []components.PageInterface{
							&components.InputTextarea{
								Label:  "Description",
								Name:   "Description",
								Rows:   4,
								Getter: getters.Key[string]("$in.Description"),
							},
						},
					},
				},
			},

			components.ContainerRow{
				Classes: "grid grid-cols-1",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.URL"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "URL",
								Name:   "URL",
								Getter: getters.Key[string]("$in.URL"),
							},
						},
					},
				},
			},

			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.ReleaseAt"),
						Children: []components.PageInterface{
							&components.InputDatetime{
								Label:    "Release At",
								Name:     "ReleaseAt",
								Required: true,
								Getter:   getters.Key[time.Time]("$in.ReleaseAt"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.ExpiryAt"),
						Children: []components.PageInterface{
							&components.InputDatetime{
								Label:    "Expiry At",
								Name:     "ExpiryAt",
								Required: false,
								Getter:   getters.Deref(getters.Key[*time.Time]("$in.ExpiryAt")),
							},
						},
					},
				},
			},
		},
	}
}

// --- Form Pages ---

func registerFormPages() {
	lago.RegistryPage.Register("announcements.AnnouncementFormFields", announcementFormFields())
	createFormName := getters.Static("announcements.AnnouncementCreateForm")
	updateFormName := getters.Static("announcements.AnnouncementUpdateForm")
	deleteFormName := getters.Static("announcements.AnnouncementDeleteForm")

	lago.RegistryPage.Register("announcements.AnnouncementCreateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "announcements.AnnouncementMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createFormName,
				ActionURL: lago.RoutePath("announcements.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Announcement]{
						Attr: getters.FormBubbling(createFormName),

						Title:    "Create Announcement",
						Subtitle: "Create a new announcement",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							announcementFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save Announcement"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("announcements.AnnouncementUpdateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "announcements.AnnouncementDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: updateFormName,
				ActionURL: lago.RoutePath("announcements.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("announcement.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Announcement]{
						Getter: getters.Key[Announcement]("announcement"),
						Attr:   getters.FormBubbling(updateFormName),

						Title:    "Edit Announcement",
						Subtitle: "Update announcement details",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							announcementFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Update Announcement"},
											&components.ButtonModalForm{
												Page:        components.Page{Roles: []string{"admin", "superuser"}},
												Label:       "Delete",
												Icon:        "trash",
												Name:        deleteFormName,
												Url:         lago.RoutePath("announcements.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("announcement.ID"))}),
												FormPostURL: lago.RoutePath("announcements.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("announcement.ID"))}),
												ModalUID:    "announcement-delete-modal",
												Classes:     "btn-error",
											},
										},
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
