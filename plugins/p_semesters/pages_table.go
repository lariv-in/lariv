package p_semesters

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("semesters.SemesterTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "semesters.SemesterMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[Semester]{
				Page:    components.Page{Key: "semesters.SemesterTableBody"},
				UID:     "semester-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Semester]]("semesters"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("semesters.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("semesters.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Code", Name: "Code", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Code")}}},
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
					{Label: "Start", Name: "Start", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Start")}}},
					{Label: "End", Name: "End", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.End")}}},
					{Label: "Active", Name: "IsActive", Children: []components.PageInterface{&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsActive")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("semesters.SemesterSelectionTable", &components.Modal{
		UID: "semester-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Semester]{
				UID:     "semester-selection-table",
				Title:   "Select semester",
				Data:    getters.Key[components.ObjectList[Semester]]("semesters"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "semesters.SemesterSelectionFilter"}},
				},
				RowAttr: getters.RowAttrSelect("SemesterID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Columns: []components.TableColumn{
					{Label: "Code", Name: "Code", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Code")}}},
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
				},
			},
		},
	})
}
