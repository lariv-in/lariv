package p_nirmancampus_sessions

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

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
								Title: "Session type",
								Children: []components.PageInterface{
									&components.FieldText{Getter: registry.PairValueFromKey(getters.Key[string]("$in.SessionType"), SessionTypeChoices)},
								},
							},
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

	lago.RegistryPage.Register("sessions.SessionDeleteForm", &components.Modal{
		UID: "session-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this session?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}

// --- Selection Tables ---
