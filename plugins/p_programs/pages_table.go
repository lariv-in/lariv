package p_programs

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("programs.ProgramTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "programs.ProgramMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[Program]{
				Page:    components.Page{Key: "programs.ProgramTableBody"},
				UID:     "program-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Program]]("programs"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("programs.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Code", Name: "Code", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Code")}}},
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
					{Label: "Standard", Name: "Standard", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Standard")}}},
					{Label: "Active", Name: "IsActive", Children: []components.PageInterface{&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsActive")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("programs.ProgramSelectionTable", &components.Modal{
		UID: "program-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Program]{
				UID:   "program-selection-table",
				Title: "Select Program",
				Data:  getters.Key[components.ObjectList[Program]]("programs"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "programs.ProgramSelectionFilter"}},
				},
				RowAttr: getters.RowAttrSelect("ProgramID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Code")),
				Columns: []components.TableColumn{
					{Label: "Code", Name: "Code", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Code")}}},
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("programs.ProgramMultiSelectionTable", &components.Modal{
		UID: "program-multi-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Program]{
				UID:     "program-multi-selection-table",
				Title:   "Select programs",
				Data:    getters.Key[components.ObjectList[Program]]("programs"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "programs.ProgramMultiSelectionFilter"}},
				},
				RowAttr: getters.RowAttrSelectMulti(
					getters.IfOrElse(getters.Key[string]("$get.target_input"), getters.Static("Programs")),
					getters.Key[uint]("$row.ID"),
					getters.Key[string]("$row.Code"),
				),
				Columns: []components.TableColumn{
					{Label: "Code", Name: "Code", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Code")}}},
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
				},
			},
		},
	})
}
