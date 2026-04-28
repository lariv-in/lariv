package p_sessions

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("sessions.ClassSessionTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "sessions.ClassSessionMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[ClassSession]{
				Page: components.Page{Key: "sessions.ClassSessionTableBody"}, UID: "class-session-table", Classes: "w-full",
				Data:    getters.Key[components.ObjectList[ClassSession]]("class_sessions"),
				Actions: []components.PageInterface{&components.TableButtonCreate{Link: lago.RoutePath("sessions.CreateRoute", nil)}},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Title", Name: "Title", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Title")}}},
					{Label: "Room", Name: "Room", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Room")}}},
					{Label: "Start", Name: "StartAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.StartAt")}}},
					{Label: "End", Name: "EndAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.EndAt")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("sessions.ClassSessionSelectionTable", &components.Modal{
		UID: "class-session-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[ClassSession]{
				UID:   "class-session-selection-table",
				Title: "Select Session",
				Data:  getters.Key[components.ObjectList[ClassSession]]("class_sessions"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "sessions.ClassSessionSelectionFilter"}},
				},
				RowAttr: getters.RowAttrSelect("SessionID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Title")),
				Columns: []components.TableColumn{
					{Label: "Title", Name: "Title", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Title")}}},
					{Label: "Room", Name: "Room", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Room")}}},
				},
			},
		},
	})
}
