package p_announcements

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_semesters"
	"gorm.io/gorm"
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

func expiryAtInputGetter() getters.Getter[time.Time] {
	return func(ctx context.Context) (time.Time, error) {
		inVal := ctx.Value(getters.ContextKeyIn) // "$in"
		if inVal == nil {
			return time.Time{}, nil
		}
		inMap, ok := inVal.(map[string]any)
		if !ok {
			return time.Time{}, nil
		}
		raw, ok := inMap["ExpiryAt"]
		if !ok || raw == nil {
			return time.Time{}, nil
		}
		switch typed := raw.(type) {
		case time.Time:
			return typed, nil
		case *time.Time:
			if typed == nil {
				return time.Time{}, nil
			}
			return *typed, nil
		default:
			return time.Time{}, nil
		}
	}
}

func expiryAtStringFromIn() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		t, err := expiryAtInputGetter()(ctx)
		if err != nil || t.IsZero() {
			return "", nil
		}
		tz, _ := ctx.Value("$tz").(*time.Location)
		if tz == nil {
			tz = components.DefaultTimeZone
		}
		return t.In(tz).Format(time.DateTime), nil
	}
}

func semesterNameFromIn() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		name, err := getters.GetterKey[string]("$in.Semester.Name")(ctx)
		if err != nil {
			return "", nil
		}
		return name, nil
	}
}

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
								Getter:   expiryAtInputGetter(),
							},
						},
					},
				},
			},

			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.SemesterID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[p_semesters.Semester]{
						Label:       "Semester",
						Name:        "SemesterID",
						Required:    true,
						Getter:      getters.GetterAssociation[p_semesters.Semester](getters.GetterKey[uint]("$in.SemesterID")),
						Url:         lago.GetterRoutePath("semesters.SelectRoute", nil),
						Display:     getters.GetterKey[string]("$in.Name"),
						Placeholder: "Select a semester...",
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

// --- Tables ---

func registerTablePages() {
	lago.RegistryPage.Register("announcements.AnnouncementTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "announcements.AnnouncementMenu"},
		},
		Children: []components.PageInterface{
			components.Environment{
				Label:   "Semester",
				Key:     getters.GetterStatic("semester"),
				Options: semestersEnvOptionsGetterForEnvironment,
			},
			&components.DataTable[Announcement]{
				UID:       "announcement-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[Announcement]]("announcements"),
				CreateUrl: lago.GetterRoutePath("announcements.CreateRoute", nil),
				OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
				})),
				FilterComponent: lago.DynamicPage{Name: "announcements.AnnouncementFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Key:   "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Title")},
						},
					},
					{
						Label: "Semester",
						Key:   "Semester",
						Children: []components.PageInterface{
							&components.FieldText{Getter: semesterNameFromRow()},
						},
					},
					{
						Label: "Release At",
						Key:   "ReleaseAt",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.ReleaseAt")},
						},
					},
					{
						Label: "Expiry At",
						Key:   "ExpiryAt",
						Children: []components.PageInterface{
							&components.FieldText{Getter: expiryAtStringFromRow()},
						},
					},
				},
			},
		},
	})
}

func semesterNameFromRow() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		name, err := getters.GetterKey[string]("$row.Semester.Name")(ctx)
		if err != nil {
			return "", nil
		}
		return name, nil
	}
}

func expiryAtStringFromRow() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		ptr, err := getters.GetterKey[*time.Time]("$row.ExpiryAt")(ctx)
		if err != nil || ptr == nil {
			return "", nil
		}
		if ptr.IsZero() {
			return "", nil
		}
		tz, _ := ctx.Value("$tz").(*time.Location)
		if tz == nil {
			tz = components.DefaultTimeZone
		}
		return ptr.In(tz).Format(time.DateTime), nil
	}
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
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Title")},
							&components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Description")},
							&components.LabelInline{
								Title:   "Semester",
								Classes: "mt-4",
								Children: []components.PageInterface{
									&components.FieldText{Getter: semesterNameFromIn()},
								},
							},
							&components.LabelInline{
								Title:   "Release At",
								Classes: "mt-4",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$in.ReleaseAt")},
								},
							},
							&components.LabelInline{
								Title:   "Expiry At",
								Classes: "mt-4",
								Children: []components.PageInterface{
									&components.FieldText{Getter: expiryAtStringFromIn()},
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
						Key:   "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Title")},
						},
					},
					{
						Label: "Semester",
						Key:   "Semester",
						Children: []components.PageInterface{
							&components.FieldText{Getter: semesterNameFromRow()},
						},
					},
					{
						Label: "Release At",
						Key:   "ReleaseAt",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.ReleaseAt")},
						},
					},
				},
			},
		},
	})
}

// semestersEnvOptionsGetterForEnvironment returns dropdown options formatted as "ID:Name".
func semestersEnvOptionsGetterForEnvironment(ctx context.Context) ([]string, error) {
	db, ok := ctx.Value("$db").(*gorm.DB)
	if !ok || db == nil {
		return nil, fmt.Errorf("semestersEnvOptionsGetterForEnvironment: missing $db in context")
	}

	var semesters []p_semesters.Semester
	if err := db.Order("start ASC").Find(&semesters).Error; err != nil {
		return nil, err
	}

	options := make([]string, 0, len(semesters))
	for _, s := range semesters {
		options = append(options, fmt.Sprintf("%d:%s", s.ID, s.Name))
	}
	return options, nil
}
