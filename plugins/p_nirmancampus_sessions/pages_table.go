package p_nirmancampus_sessions

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerFilterPages() {
	lago.RegistryPage.Register("sessions.SessionFilter", &components.FormComponent[AdmissionSession]{
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

	lago.RegistryPage.Register("sessions.sessionselectionFilter", &components.FormComponent[AdmissionSession]{
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

func registerTablePages() {
	lago.RegistryPage.Register("sessions.SessionTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sessions.SessionMenu"},
		},
		Children: []components.PageInterface{
			&components.FieldTitle{Getter: getters.Static("All Sessions"), Classes: "mb-4"},
			&components.DataTable[AdmissionSession]{
				Page:    components.Page{Key: "sessions.SessionTableBodyAdmission"},
				UID:     "session-table-admission",
				Title:   "Admission sessions",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[AdmissionSession]]("sessions"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "sessions.SessionFilter"}},
					&components.TableButtonCreate{Link: lago.RoutePath("sessions.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: admissionSessionTableColumns(),
			},
		},
	})
}

func admissionSessionTableColumns() []components.TableColumn {
	return []components.TableColumn{
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
				&components.FieldDate{Getter: getters.Key[time.Time]("$row.Start")},
			},
		},
		{
			Label: "End",
			Name:  "End",
			Children: []components.PageInterface{
				&components.FieldDate{Getter: getters.Key[time.Time]("$row.End")},
			},
		},
		{
			Label: "Active",
			Name:  "IsActive",
			Children: []components.PageInterface{
				&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsActive")},
			},
		},
	}
}

// --- Detail & Delete ---
