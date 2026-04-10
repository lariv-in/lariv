package p_nirmancampus_courses

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerFilterPages() {
	lago.RegistryPage.Register("courses.CourseFilter", &components.FormComponent[Course]{
		Attr: getters.FormBoostedGet(lago.RoutePath("courses.DefaultRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.Key[string]("$get.Code")},
			&components.InputText{Label: "Type", Name: "CourseType", Getter: getters.Key[string]("$get.CourseType")},
			&components.InputTernary{
				Label:      "Active",
				Name:       "IsActive",
				TrueLabel:  "Active Only",
				FalseLabel: "Inactive Only",
				NoneLabel:  "All",
				Getter:     getters.Key[bool]("$get.IsActive"),
			},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply Filters"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("courses.CourseSelectionFilter", &components.FormComponent[Course]{
		Attr: getters.FormBoostedGet(lago.RoutePath("courses.SelectRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.Key[string]("$get.Code")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("courses.CourseMultiSelectionFilter", &components.FormComponent[Course]{
		Attr: getters.FormBoostedGet(lago.RoutePath("courses.MultiSelectRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{Hidden: true, Name: "target_input", Getter: getters.Key[string]("$get.target_input")},
			&components.InputText{Hidden: true, Name: "pool_course_ids", Getter: getters.Key[string]("$get.pool_course_ids")},
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.Key[string]("$get.Code")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})
}

func registerTablePages() {
	lago.RegistryPage.Register("courses.CourseTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Course]{
				Page:    components.Page{Key: "courses.CourseTableBody"},
				UID:     "course-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Course]]("courses"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "courses.CourseFilter"}},
					&components.ButtonModalForm{
						Page:        components.Page{Roles: []string{"admin", "superuser"}},
						Name:        getters.Static("courses.CourseCreateForm"),
						Url:         lago.RoutePath("courses.CreateRoute", nil),
						FormPostURL: lago.RoutePath("courses.CreateRoute", nil),
						ModalUID:    "courses-create-modal",
						Icon:        "plus",
						Classes:     "btn-square btn-outline btn-sm",
						Attr:        getters.ModalRefreshList(getters.Static(""), getters.Static("#course-table")),
					},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
					{Label: "Code", Name: "Code", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Code")},
					}},
					{Label: "Type", Name: "CourseType", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.CourseType")},
					}},
					{Label: "Active", Name: "IsActive", Children: []components.PageInterface{
						&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsActive")},
					}},
				},
			},
		},
	})
}

func registerSelectionPages() {
	lago.RegistryPage.Register("courses.CourseSelectionTable", &components.Modal{
		UID: "course-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Course]{
				UID:     "course-selection-table",
				Title:   "Select Course",
				Data:    getters.Key[components.ObjectList[Course]]("courses"),
				RowAttr: getters.RowAttrSelect("CourseID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "courses.CourseSelectionFilter"}},
					&components.ButtonModalForm{
						Page:        components.Page{Roles: []string{"admin", "superuser"}},
						Name:        getters.Static("courses.CourseCreateForm"),
						Url:         lago.RoutePath("courses.CreateRoute", nil),
						FormPostURL: lago.RoutePath("courses.CreateRoute", nil),
						ModalUID:    "courses-create-modal",
						Icon:        "plus",
						Classes:     "btn-square btn-outline btn-sm",
						Attr:        getters.ModalRefreshList(getters.Static(""), getters.Static("#course-selection-table")),
					},
				},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
					{Label: "Code", Name: "Code", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Code")},
					}},
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseMultiSelectionTable", &components.Modal{
		UID: "course-multi-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Course]{
				UID:   "course-multi-selection-table",
				Title: "Select Courses",
				Data:  getters.Key[components.ObjectList[Course]]("courses"),
				RowAttr: getters.RowAttrSelectMulti(
					getters.IfOrElse(
						getters.Key[string]("$get.target_input"),
						getters.Static("Courses"),
					),
					getters.Key[uint]("$row.ID"),
					getters.Key[string]("$row.Name"),
				),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "courses.CourseMultiSelectionFilter"}},
					&components.ButtonModalForm{
						Page:        components.Page{Roles: []string{"admin", "superuser"}},
						Name:        getters.Static("courses.CourseCreateForm"),
						Url:         lago.RoutePath("courses.CreateRoute", nil),
						FormPostURL: lago.RoutePath("courses.CreateRoute", nil),
						ModalUID:    "courses-create-modal",
						Icon:        "plus",
						Classes:     "btn-square btn-outline btn-sm",
						Attr:        getters.ModalRefreshList(getters.Static(""), getters.Static("#course-multi-selection-table")),
					},
				},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
					{Label: "Code", Name: "Code", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Code")},
					}},
				},
			},
		},
	})
}
