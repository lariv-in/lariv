package p_nirmancampus_announcements

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
		Title: getters.Static("Announcements"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Announcements"),
				Url:   lago.RoutePath("announcements.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("announcements.AnnouncementDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Announcement: %s", getters.Any(getters.Key[string]("announcement.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all Announcements"),
			Url:   lago.RoutePath("announcements.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Announcement Detail"),
				Url: lago.RoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("announcement.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit Announcement"),
				Url: lago.RoutePath("announcements.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("announcement.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Delete Announcement"),
				Url: lago.RoutePath("announcements.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("announcement.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("announcements.AnnouncementFilter", &components.FormComponent[Announcement]{
		Url:    lago.RoutePath("announcements.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Title",
				Name:   "Title",
				Getter: getters.Key[string]("$get.Title"),
			},
			&components.InputText{
				Label:  "Description",
				Name:   "Description",
				Getter: getters.Key[string]("$get.Description"),
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
		Url:    lago.RoutePath("announcements.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Title",
				Name:   "Title",
				Getter: getters.Key[string]("$get.Title"),
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

	lago.RegistryPage.Register("announcements.AnnouncementCreateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "announcements.AnnouncementMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Announcement]{
				Url:      lago.RoutePath("announcements.CreateRoute", nil),
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
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "announcements.AnnouncementDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Announcement]{
				Getter: getters.Key[Announcement]("announcement"),
				Url: lago.RoutePath("announcements.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$in.ID")),
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
		role, err := getters.Key[string]("$role")(ctx)
		if err != nil {
			return "", err
		}
		if role == "superuser" || role == "admin" {
			return lago.RoutePath("announcements.CreateRoute", nil)(ctx)
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
				Page:    components.Page{Key: "announcements.AnnouncementTableBody"},
				UID:     "announcement-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Announcement]]("announcements"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{
						Child: lago.DynamicPage{Name: "announcements.AnnouncementFilter"},
					},
					&components.TableButtonCreate{
						Link: announcementCreateUrlGetter(),
						Page: components.Page{Roles: []string{"admin", "superuser"}},
					},
				},
				OnClick: getters.NavigateGetter(lago.RoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$row.ID")),
				})),
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Title")},
						},
					},
					{
						Label: "URL",
						Name:  "URL",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.URL")},
						},
					},
					{
						Label: "Release At",
						Name:  "ReleaseAt",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.ReleaseAt")},
						},
					},
					{
						Label: "Expiry At",
						Name:  "ExpiryAt",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Deref(getters.Key[*time.Time]("$row.ExpiryAt"))},
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
				Getter: getters.Key[Announcement]("announcement"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page: components.Page{Key: "announcements.AnnouncementDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Page:   components.Page{Key: "announcements.AnnouncementDetailTitle"},
								Getter: getters.Key[string]("$in.Title"),
							},
							&components.LabelInline{
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

	lago.RegistryPage.Register("announcements.AnnouncementDeleteForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "announcements.AnnouncementDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this announcement?",
				CancelUrl: lago.RoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("announcement.ID")),
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
				Data: getters.Key[components.ObjectList[Announcement]]("announcements"),
				OnClick: getters.Select("AnnouncementID",
					getters.Key[uint]("$row.ID"),
					getters.Key[string]("$row.Title"),
				),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{
						Child: lago.DynamicPage{Name: "announcements.AnnouncementSelectionFilter"},
						Page:  components.Page{Roles: []string{"admin", "superuser"}},
					},
				},
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Title")},
						},
					},
					{
						Label: "Release At",
						Name:  "ReleaseAt",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.ReleaseAt")},
						},
					},
				},
			},
		},
	})
}
