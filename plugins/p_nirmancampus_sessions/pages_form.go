package p_nirmancampus_sessions

import (
	"context"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

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
							&components.InputDate{
								Label:    "Start Date",
								Name:     "Start",
								Required: true,
								Getter:   getters.Key[time.Time]("$in.Start"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.End"),
						Children: []components.PageInterface{
							&components.InputDate{
								Label:    "End Date",
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
				Name:      getters.Static("sessions.SessionCreateForm"),
				ActionURL: lago.RoutePath("sessions.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Session]{
						Attr: getters.FormBubbling(getters.Static("sessions.SessionCreateForm")),

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
				Name: getters.Static("sessions.SessionUpdateForm"),
				ActionURL: lago.RoutePath("sessions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("session.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Session]{
						Getter: getters.Key[Session]("session"),
						Attr:   getters.FormBubbling(getters.Static("sessions.SessionUpdateForm")),

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
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save Session"},
											&components.ButtonModalForm{
												Label:       "Delete",
												Icon:        "trash",
												Name:        getters.Static("sessions.SessionDeleteForm"),
												Url:         lago.RoutePath("sessions.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("session.ID"))}),
												FormPostURL: lago.RoutePath("sessions.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("session.ID"))}),
												ModalUID:    "session-delete-modal",
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

// --- Tables ---
