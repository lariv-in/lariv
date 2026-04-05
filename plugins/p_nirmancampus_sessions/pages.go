package p_nirmancampus_sessions

import (
	"context"
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
	lago.RegistryPage.Register("sessions.SessionMenu", &components.SidebarMenu{
		Title: getters.Static("sessions"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All sessions"),
				Url:   lago.RoutePath("sessions.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("sessions.SessionDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Session: %s", getters.Any(getters.Key[string]("session.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all sessions"),
			Url:   lago.RoutePath("sessions.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Session Detail"),
				Url: lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("session.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit Session"),
				Url: lago.RoutePath("sessions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("session.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("sessions.SessionFilter", &components.FormComponent[Session]{
		Attr: getters.FormBoostedGet(lago.RoutePath("sessions.DefaultRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.Key[string]("$get.Code"),
			},
			&components.InputTernary{
				Label:      "Active",
				Name:       "IsActiveFilter",
				TrueLabel:  "Active Only",
				FalseLabel: "Inactive Only",
				NoneLabel:  "All",
				// Intentionally omit Getter: we want the default selection to be "All".
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

	lago.RegistryPage.Register("sessions.sessionselectionFilter", &components.FormComponent[Session]{
		Attr: getters.FormBoostedGet(lago.RoutePath("sessions.SelectRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.Key[string]("$get.Code"),
			},
			&components.InputTernary{
				Label:      "Active",
				Name:       "IsActiveFilter",
				TrueLabel:  "Active Only",
				FalseLabel: "Inactive Only",
				NoneLabel:  "All",
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

// --- Form Fields ---

func isActiveGetter() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		in := ctx.Value(getters.ContextKeyIn)
		if in == nil {
			// Create forms don't provide $in; default the checkbox to "true".
			return true, nil
		}
		m, ok := in.(map[string]any)
		if !ok {
			return true, nil
		}
		raw, ok := m["IsActive"]
		if !ok || raw == nil {
			return true, nil
		}
		v, ok := raw.(bool)
		if !ok {
			return true, nil
		}
		return v, nil
	}
}

func sessionCodeInputGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		// Code is auto-generated in BeforeSave when blank, so keep it empty on create.
		return getters.Key[string]("$in.Code")(ctx)
	}
}

func sessionFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "sessions.SessionFormFieldsBody",
		},
		Children: []components.PageInterface{
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Name",
								Name:     "Name",
								Required: true,
								Getter:   getters.Key[string]("$in.Name"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Code"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Code",
								Name:   "Code",
								Getter: sessionCodeInputGetter(),
							},
						},
					},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Start"),
						Children: []components.PageInterface{
							&components.InputDatetime{
								Label:    "Start Date & Time",
								Name:     "Start",
								Required: true,
								Getter:   getters.Key[time.Time]("$in.Start"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.End"),
						Children: []components.PageInterface{
							&components.InputDatetime{
								Label:    "End Date & Time",
								Name:     "End",
								Required: true,
								Getter:   getters.Key[time.Time]("$in.End"),
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.IsActive"),
				Children: []components.PageInterface{
					&components.InputCheckbox{
						Label:    "Active",
						Name:     "IsActive",
						Getter:   isActiveGetter(),
						Required: false,
					},
				},
			},
		},
	}
}

// --- Form Pages ---

func registerFormPages() {
	lago.RegistryPage.Register("sessions.SessionFormFields", sessionFormFields())

	lago.RegistryPage.Register("sessions.SessionCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.SessionMenu"},
		},
		Children: []components.PageInterface{
						&components.FormListenBoostedPost{
				ActionURL: lago.RoutePath("sessions.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Session]{
				Attr: getters.FormBubbling(nil),


				Title:    "Create Session",
				Subtitle: "Create a new session",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					sessionFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Session"},
				},
				},
			},
		},
		},
	})

	lago.RegistryPage.Register("sessions.SessionUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.SessionDetailMenu"},
		},
		Children: []components.PageInterface{
						&components.FormListenBoostedPost{
				ActionURL: lago.RoutePath("sessions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("session.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Session]{
				Getter: getters.Key[Session]("session"),
				Attr: getters.FormBubbling(nil),


				Title:    "Edit Session",
				Subtitle: "Update session details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					sessionFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ContainerRow{
						Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
						Children: []components.PageInterface{
							&components.ButtonModalForm{
								Label:       "Delete",
								Icon:        "trash",
								Url:         lago.RoutePath("sessions.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("session.ID"))}),
								FormPostURL: lago.RoutePath("sessions.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("session.ID"))}),
								ModalUID:    "session-delete-modal",
								Classes:     "btn-outline btn-error btn-sm",
							},
							&components.ContainerRow{
								Classes: "flex justify-end gap-2",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save Session"},
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

// --- Tables ---

func registerTablePages() {
	lago.RegistryPage.Register("sessions.SessionTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.SessionMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Session]{
				Page:    components.Page{Key: "sessions.SessionTableBody"},
				UID:     "session-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Session]]("sessions"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "sessions.SessionFilter"}},
					&components.TableButtonCreate{Link: lago.RoutePath("sessions.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Code")},
						},
					},
					{
						Label: "Start",
						Name:  "Start",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Start")},
						},
					},
					{
						Label: "End",
						Name:  "End",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.End")},
						},
					},
					{
						Label: "Active",
						Name:  "IsActive",
						Children: []components.PageInterface{
							&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsActive")},
						},
					},
				},
			},
		},
	})
}

// --- Detail & Delete ---

func registerDetailPages() {
	lago.RegistryPage.Register("sessions.SessionDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.SessionDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Session]{
				Getter: getters.Key[Session]("session"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "sessions.SessionDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Code")},
							&components.LabelInline{
								Title: "Active",
								Children: []components.PageInterface{
									&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsActive")},
								},
							},
							&components.LabelInline{
								Title: "Start",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.Start")},
								},
							},
							&components.LabelInline{
								Title: "End",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.End")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("sessions.SessionDeleteForm", &components.Modal{
		UID: "session-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this session?",
				Attr: getters.FormBubbling(nil),
			},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("sessions.sessionselectionTable", &components.Modal{
		UID: "session-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Session]{
				Page:    components.Page{Key: "sessions.sessionselectionTableBody"},
				UID:     "session-selection-table",
				Title:   "Select Session",
				Data:    getters.Key[components.ObjectList[Session]]("sessions"),
				RowAttr: getters.RowAttrSelect("SessionID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "sessions.sessionselectionFilter"}},
				},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Code")},
						},
					},
					{
						Label: "Start",
						Name:  "Start",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Start")},
						},
					},
					{
						Label: "Active",
						Name:  "IsActive",
						Children: []components.PageInterface{
							&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsActive")},
						},
					},
				},
			},
		},
	})
}
