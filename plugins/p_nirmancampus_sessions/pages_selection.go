package p_nirmancampus_sessions

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

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
