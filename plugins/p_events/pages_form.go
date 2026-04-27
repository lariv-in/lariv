package p_events

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_semesters"
)

func schoolEventFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "events.SchoolEventFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Title"),
				Children: []components.PageInterface{
					&components.InputText{Label: "Title", Name: "Title", Required: true, Getter: getters.Key[string]("$in.Title")},
				},
			},
			&components.InputTextarea{Label: "Description", Name: "Description", Rows: 4, Getter: getters.Key[string]("$in.Description")},
			&components.ContainerError{
				Error: getters.Key[error]("$error.StartsAt"),
				Children: []components.PageInterface{
					&components.InputDatetime{Label: "Starts at", Name: "StartsAt", Required: true, Getter: getters.Key[time.Time]("$in.StartsAt")},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.EndsAt"),
				Children: []components.PageInterface{
					&components.InputDatetime{Label: "Ends at (optional)", Name: "EndsAt", Required: false, Getter: getters.Deref(getters.Key[*time.Time]("$in.EndsAt"))},
				},
			},
			&components.InputCheckbox{Label: "Universal", Name: "IsUniversal", Getter: getters.Key[bool]("$in.IsUniversal")},
			&components.InputCheckbox{Label: "Active", Name: "IsActive", Getter: getters.Key[bool]("$in.IsActive")},
			&components.InputForeignKey[p_semesters.Semester]{Label: "Semester (optional)", Name: "SemesterID", Required: false, Url: lago.RoutePath("semesters.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "Semester…", Getter: getters.Association[p_semesters.Semester](getters.Deref(getters.Key[*uint]("$in.SemesterID")))},
		},
	}
}

func registerFormPages() {
	deleteFormName := getters.Static("events.SchoolEventDeleteForm")
	lago.RegistryPage.Register("events.SchoolEventCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "events.SchoolEventMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("events.SchoolEventCreateForm"), ActionURL: lago.RoutePath("events.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[SchoolEvent]{
						Attr: getters.FormBubbling(getters.Static("events.SchoolEventCreateForm")), Title: "Create event",
						ChildrenInput:  []components.PageInterface{schoolEventFormFields()},
						ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save"}},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("events.SchoolEventUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "events.SchoolEventDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("events.SchoolEventUpdateForm"),
				ActionURL: lago.RoutePath("events.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("school_event.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[SchoolEvent]{
						Getter:        getters.Key[SchoolEvent]("school_event"),
						Attr:          getters.FormBubbling(getters.Static("events.SchoolEventUpdateForm")),
						Title:         "Edit event",
						ChildrenInput: []components.PageInterface{schoolEventFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label:       "Delete",
										Icon:        "trash",
										Name:        deleteFormName,
										Url:         lago.RoutePath("events.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("school_event.ID"))}),
										FormPostURL: lago.RoutePath("events.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("school_event.ID"))}),
										ModalUID:    "school-event-delete-modal",
										Classes:     "btn-error",
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
