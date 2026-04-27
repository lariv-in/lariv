package p_syllabus

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("syllabus.SyllabusTopicTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "syllabus.SyllabusTopicMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[SyllabusTopic]{
				Page:    components.Page{Key: "syllabus.SyllabusTopicTableBody"},
				UID:     "syllabus-topic-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[SyllabusTopic]]("syllabus_topics"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("syllabus.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("syllabus.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$row.ID")),
				})),
				Columns: []components.TableColumn{
					{Label: "Course ID", Name: "CourseID", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.CourseID")))},
					}},
					{Label: "Title", Name: "Title", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Title")},
					}},
					{Label: "Sort", Name: "SortOrder", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.SortOrder")))},
					}},
				},
			},
		},
	})

	lago.RegistryPage.Register("syllabus.SyllabusTopicMultiSelectionTable", &components.Modal{
		UID: "syllabus-topic-multi-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[SyllabusTopic]{
				UID:     "syllabus-topic-multi-selection-table",
				Title:   "Select syllabus topics",
				Data:    getters.Key[components.ObjectList[SyllabusTopic]]("syllabus_topics"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "syllabus.SyllabusTopicMultiSelectionFilter"}},
				},
				RowAttr: getters.RowAttrSelectMulti(
					getters.IfOrElse(getters.Key[string]("$get.target_input"), getters.Static("Topics")),
					getters.Key[uint]("$row.ID"),
					getters.Key[string]("$row.Title"),
				),
				Columns: []components.TableColumn{
					{Label: "Course ID", Name: "CourseID", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.CourseID")))},
					}},
					{Label: "Title", Name: "Title", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Title")},
					}},
					{Label: "Sort", Name: "SortOrder", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.SortOrder")))},
					}},
				},
			},
		},
	})
}
