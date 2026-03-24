package p_announcements

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
}

// --- Menus ---

func registerMenuPages() {
	lago.RegistryPage.Register("announcements.AnnouncementMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Announcements"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Announcements"),
				Url:   lago.GetterRoutePath("announcements.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("announcements.AnnouncementDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Announcement: %s", getters.GetterAny(getters.GetterKey[string]("announcement.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Announcements"),
			Url:   lago.GetterRoutePath("announcements.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Announcement Detail"),
				Url: lago.GetterRoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("announcement.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Announcement"),
				Url: lago.GetterRoutePath("announcements.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("announcement.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Announcement"),
				Url: lago.GetterRoutePath("announcements.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("announcement.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("announcements.AnnouncementFilter", &components.FormComponent[Announcement]{
		Url:    lago.GetterRoutePath("announcements.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Title",
				Name:   "Title",
				Getter: getters.GetterKey[string]("$get.Title"),
			},
			&components.InputText{
				Label:  "Description",
				Name:   "Description",
				Getter: getters.GetterKey[string]("$get.Description"),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply Filters"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})

	lago.RegistryPage.Register("announcements.AnnouncementSelectionFilter", &components.FormComponent[Announcement]{
		Url:    lago.GetterRoutePath("announcements.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Title",
				Name:   "Title",
				Getter: getters.GetterKey[string]("$get.Title"),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})
}

// --- Form Fields / Helpers ---

func announcementFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "announcements.AnnouncementFormFieldsBody"},
		Children: []components.PageInterface{
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Title"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Title",
								Name:     "Title",
								Required: true,
								Getter:   getters.GetterKey[string]("$in.Title"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Description"),
						Children: []components.PageInterface{
							&components.InputTextarea{
								Label:    "Description",
								Name:     "Description",
								Rows:     4,
								Required: true,
								Getter:   getters.GetterKey[string]("$in.Description"),
							},
						},
					},
				},
			},

			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.ReleaseAt"),
						Children: []components.PageInterface{
							&components.InputDatetime{
								Label:    "Release At",
								Name:     "ReleaseAt",
								Required: true,
								Getter:   getters.GetterKey[time.Time]("$in.ReleaseAt"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.ExpiryAt"),
						Children: []components.PageInterface{
							&components.InputDatetime{
								Label:    "Expiry At",
								Name:     "ExpiryAt",
								Required: false,
								Getter:   getters.GetterDeref(getters.GetterKey[*time.Time]("$in.ExpiryAt")),
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

	lago.RegistryPage.Register("announcements.AnnouncementCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "announcements.AnnouncementMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Announcement]{
				Url:      lago.GetterRoutePath("announcements.CreateRoute", nil),
				Method:   http.MethodPost,
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
	})

	lago.RegistryPage.Register("announcements.AnnouncementUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "announcements.AnnouncementDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Announcement]{
				Getter: getters.GetterKey[Announcement]("announcement"),
				Url: lago.GetterRoutePath("announcements.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Announcement",
				Subtitle: "Update announcement details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					announcementFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Update Announcement"},
				},
			},
		},
	})
}

func announcementCreateUrlGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		role, err := getters.GetterKey[string]("$role")(ctx)
		if err != nil {
			return "", err
		}
		if role == "superuser" || role == "nirmancampus_admin" {
			return lago.GetterRoutePath("announcements.CreateRoute", nil)(ctx)
		}
		return "", fmt.Errorf("you do not have permission to do this action")
	}
}

// --- Tables ---

func registerTablePages() {
	lago.RegistryPage.Register("announcements.AnnouncementTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "announcements.AnnouncementMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Announcement]{
				Page:      components.Page{Key: "announcements.AnnouncementTableBody"},
				UID:       "announcement-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[Announcement]]("announcements"),
				CreateUrl: announcementCreateUrlGetter(),
				OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
				})),
				FilterComponent: lago.DynamicPage{Name: "announcements.AnnouncementFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Title")},
						},
					},
					{
						Label: "Release At",
						Name:  "ReleaseAt",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.ReleaseAt")},
						},
					},
					{
						Label: "Expiry At",
						Name:  "ExpiryAt",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.GetterDeref(getters.GetterKey[*time.Time]("$row.ExpiryAt"))},
						},
					},
				},
			},
		},
	})
}

// --- Detail & Delete ---

func registerDetailPages() {
	lago.RegistryPage.Register("announcements.AnnouncementDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "announcements.AnnouncementDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Announcement]{
				Getter: getters.GetterKey[Announcement]("announcement"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "announcements.AnnouncementDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Page:   components.Page{Key: "announcements.AnnouncementDetailTitle"},
								Getter: getters.GetterKey[string]("$in.Title"),
							},
							&components.LabelInline{
								Title: "Description",
								Children: []components.PageInterface{
									&components.FieldMarkdown{Getter: getters.GetterKey[string]("$in.Description")},
								},
							},
							&components.LabelInline{
								Title: "Release At",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$in.ReleaseAt")},
								},
							},
							&components.LabelInline{
								Title: "Expiry At",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.GetterDeref(getters.GetterKey[*time.Time]("$in.ExpiryAt"))},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("announcements.AnnouncementDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "announcements.AnnouncementDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this announcement?",
				CancelUrl: lago.GetterRoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("announcement.ID")),
				}),
			},
		},
	})
}

// --- Selection ---

func registerSelectionPages() {
	lago.RegistryPage.Register("announcements.AnnouncementSelectionTable", &components.Modal{
		UID:   "announcement-selection-modal",
		Title: "Select Announcement",
		Children: []components.PageInterface{
			&components.DataTable[Announcement]{
				Page: components.Page{Key: "announcements.AnnouncementSelectionTableBody"},
				UID:  "announcement-selection-table",
				Data: getters.GetterKey[components.ObjectList[Announcement]]("announcements"),
				OnClick: getters.GetterSelect("AnnouncementID",
					getters.GetterKey[uint]("$row.ID"),
					getters.GetterKey[string]("$row.Title"),
				),
				FilterComponent: lago.DynamicPage{Name: "announcements.AnnouncementSelectionFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Title")},
						},
					},
					{
						Label: "Release At",
						Name:  "ReleaseAt",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.ReleaseAt")},
						},
					},
				},
			},
		},
	})
}
