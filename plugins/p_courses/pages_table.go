package p_courses

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("courses.CourseTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "courses.CourseMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[Course]{
				Page:    components.Page{Key: "courses.CourseTableBody"},
				UID:     "course-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Course]]("courses"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("courses.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))}),
				),
				Columns: []components.TableColumn{
					{Label: "Code", Name: "Code", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Code")}}},
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
					{Label: "Active", Name: "IsActive", Children: []components.PageInterface{&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsActive")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseSelectionTable", &components.Modal{
		UID: "course-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Course]{
				UID:   "course-selection-table",
				Title: "Select Course",
				Data:  getters.Key[components.ObjectList[Course]]("courses"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "courses.CourseSelectionFilter"}},
				},
				RowAttr: getters.RowAttrSelect("CourseID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Code")),
				Columns: []components.TableColumn{
					{Label: "Code", Name: "Code", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Code")}}},
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
				},
			},
		},
	})
}
