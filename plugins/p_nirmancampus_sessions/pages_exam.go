package p_nirmancampus_sessions

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerExamPages() {
	registerExamMenuPages()
	registerExamFormPages()
	registerExamDetailPages()
}

func registerExamMenuPages() {
	lago.RegistryPage.Register("sessions.ExamDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Exam: %s", getters.Any(getters.Key[string]("exam_session.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all sessions"),
			Url:   lago.RoutePath("sessions.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("sessions.ExamDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("exam_session.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("sessions.ExamUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("exam_session.ID")),
				}),
			},
		},
	})
}

func examSessionFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "sessions.ExamFormFieldsBody",
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

func registerExamFormPages() {
	lago.RegistryPage.Register("sessions.ExamFormFields", examSessionFormFields())

	lago.RegistryPage.Register("sessions.ExamCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.SessionMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("sessions.ExamCreateForm"),
				ActionURL: lago.RoutePath("sessions.ExamCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[ExamSession]{
						Attr: getters.FormBubbling(getters.Static("sessions.ExamCreateForm")),

						Title:    "Create exam session",
						Subtitle: "Create a new exam session",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							examSessionFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("sessions.ExamUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.ExamDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("sessions.ExamUpdateForm"),
				ActionURL: lago.RoutePath("sessions.ExamUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("exam_session.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[ExamSession]{
						Getter: getters.Key[ExamSession]("exam_session"),
						Attr:   getters.FormBubbling(getters.Static("sessions.ExamUpdateForm")),

						Title:    "Edit exam session",
						Subtitle: "Update exam session details",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							examSessionFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save"},
											&components.ButtonModalForm{
												Label:       "Delete",
												Icon:        "trash",
												Name:        getters.Static("sessions.ExamDeleteForm"),
												Url:         lago.RoutePath("sessions.ExamDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("exam_session.ID"))}),
												FormPostURL: lago.RoutePath("sessions.ExamDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("exam_session.ID"))}),
												ModalUID:    "exam-session-delete-modal",
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

func registerExamDetailPages() {
	lago.RegistryPage.Register("sessions.ExamDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.ExamDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[ExamSession]{
				Getter: getters.Key[ExamSession]("exam_session"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "sessions.ExamDetailContent"},
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
									&components.FieldDate{Getter: getters.Key[time.Time]("$in.Start")},
								},
							},
							&components.LabelInline{
								Title: "End",
								Children: []components.PageInterface{
									&components.FieldDate{Getter: getters.Key[time.Time]("$in.End")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("sessions.ExamDeleteForm", &components.Modal{
		UID: "exam-session-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this exam session?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
