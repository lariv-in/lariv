package p_nirmancampus_announcements

import (
	"fmt"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerFilterPages() {
	lago.RegistryPage.Register("announcements.AnnouncementFilter", &components.FormComponent[Announcement]{
		Attr: getters.FormBoostedGet(lago.RoutePath("announcements.DefaultRoute", nil)),

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
		Attr: getters.FormBoostedGet(lago.RoutePath("announcements.SelectRoute", nil)),

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

func registerTablePages() {
	create := lago.RoutePath("announcements.CreateRoute", nil)
	announcementTableCreateLink := getters.Match(getters.Key[string]("$role"), map[string]getters.Getter[string]{
		"superuser": create,
		"admin":     create,
	}, getters.Static(fmt.Errorf("you do not have permission to do this action")))
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
						Link: announcementTableCreateLink,
						Page: components.Page{Roles: []string{"admin", "superuser"}},
					},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
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

func registerSelectionPages() {
	lago.RegistryPage.Register("announcements.AnnouncementSelectionTable", &components.Modal{
		UID: "announcement-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Announcement]{
				Page:  components.Page{Key: "announcements.AnnouncementSelectionTableBody"},
				UID:   "announcement-selection-table",
				Title: "Select Announcement",
				Data:  getters.Key[components.ObjectList[Announcement]]("announcements"),
				RowAttr: getters.RowAttrSelect("AnnouncementID",
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
