package p_teachers

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("teachers.TeacherTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "teachers.TeacherMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Teacher]{
				Page:    components.Page{Key: "teachers.TeacherTableBody"},
				UID:     "teacher-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Teacher]]("teachers"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("teachers.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{Label: "Code", Name: "Code", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Code")}}},
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
					{Label: "Email", Name: "Email", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Email")}}},
					{Label: "Phone", Name: "Phone", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Phone")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("teachers.TeacherSelectionTable", &components.Modal{
		UID: "teacher-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Teacher]{
				UID:   "teacher-selection-table",
				Title: "Select Teacher",
				Data:  getters.Key[components.ObjectList[Teacher]]("teachers"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "teachers.TeacherSelectionFilter"}},
				},
				RowAttr: getters.RowAttrSelect("TeacherID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Code")),
				Columns: []components.TableColumn{
					{Label: "Code", Name: "Code", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Code")}}},
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("teachers.TeacherMultiSelectionTable", &components.Modal{
		UID: "teacher-multi-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Teacher]{
				UID:     "teacher-multi-selection-table",
				Title:   "Select teachers",
				Data:    getters.Key[components.ObjectList[Teacher]]("teachers"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "teachers.TeacherMultiSelectionFilter"}},
				},
				RowAttr: getters.RowAttrSelectMulti(
					getters.IfOrElse(getters.Key[string]("$get.target_input"), getters.Static("Teachers")),
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
