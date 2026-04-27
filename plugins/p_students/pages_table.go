package p_students

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("students.StudentTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Student]{
				Page:    components.Page{Key: "students.StudentTableBody"},
				UID:     "student-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Student]]("students"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("students.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Student No",
						Name:  "StudentNo",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.StudentNo")},
						},
					},
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						},
					},
					{
						Label: "Email",
						Name:  "Email",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Email")},
						},
					},
					{
						Label: "Phone",
						Name:  "Phone",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Phone")},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("students.StudentSelectionTable", &components.Modal{
		UID: "student-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Student]{
				UID:     "student-selection-table",
				Title:   "Select Student",
				Data:    getters.Key[components.ObjectList[Student]]("students"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "students.StudentSelectionFilter"}},
				},
				RowAttr: getters.RowAttrSelect("StudentID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.StudentNo")),
				Columns: []components.TableColumn{
					{Label: "Student No", Name: "StudentNo", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.StudentNo")}}},
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("students.StudentMultiSelectionTable", &components.Modal{
		UID: "student-multi-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Student]{
				UID:     "student-multi-selection-table",
				Title:   "Select students",
				Data:    getters.Key[components.ObjectList[Student]]("students"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "students.StudentMultiSelectionFilter"}},
				},
				RowAttr: getters.RowAttrSelectMulti(
					getters.IfOrElse(getters.Key[string]("$get.target_input"), getters.Static("Students")),
					getters.Key[uint]("$row.ID"),
					getters.Key[string]("$row.StudentNo"),
				),
				Columns: []components.TableColumn{
					{Label: "Student No", Name: "StudentNo", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.StudentNo")}}},
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
				},
			},
		},
	})
}
