package p_nirmancampus_sessions

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

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
								// Empty on create; auto-generated in BeforeSave when blank.
								Getter: getters.Key[string]("$in.Code"),
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
						Getter:   getters.Key[bool]("$in.IsActive"),
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
