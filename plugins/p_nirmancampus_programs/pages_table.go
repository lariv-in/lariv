package p_nirmancampus_programs

import (
	"fmt"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

func universityFilterSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:   "University",
		Name:    "University",
		Choices: getters.Static(UniversityChoices),
		Getter:  registry.PairFromGetter(getters.Key[string]("$get.University"), UniversityChoices),
	}
}

func programTypeFilterSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:   "Program type",
		Name:    "ProgramType",
		Choices: getters.Static(programTypeChoices),
		Getter:  registry.PairFromGetter(getters.Key[string]("$get.ProgramType"), programTypeChoices),
	}
}

func registerFilterPages() {
	lago.RegistryPage.Register("programs.ProgramFilter", &components.FormComponent[Program]{
		Attr: getters.FormBoostedGet(lago.RoutePath("programs.DefaultRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.Key[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.Key[string]("$get.Code"),
			},
			universityFilterSelect(),
			programTypeFilterSelect(),
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

	lago.RegistryPage.Register("programs.ProgramSelectionFilter", &components.FormComponent[Program]{
		Attr: getters.FormBoostedGet(lago.RoutePath("programs.SelectRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.Key[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.Key[string]("$get.Code"),
			},
			universityFilterSelect(),
			programTypeFilterSelect(),
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
	lago.RegistryPage.Register("programs.ProgramTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Program]{
				Page:    components.Page{Key: "programs.ProgramTableBody"},
				UID:     "program-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Program]]("programs"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "programs.ProgramFilter"}},
					&components.TableButtonCreate{Link: getters.Match(getters.Key[string]("$role"), map[string]getters.Getter[string]{
						"superuser": lago.RoutePath("programs.CreateRoute", nil),
						"admin":     lago.RoutePath("programs.CreateRoute", nil),
					}, getters.Static(fmt.Errorf("you do not have permission to do this action")))},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
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
						Label: "University",
						Name:  "University",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.University")},
						},
					},
					{
						Label: "Program type",
						Name:  "ProgramType",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.ProgramType")},
						},
					},
					{
						Label: "Description",
						Name:  "Description",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Description")},
						},
					},
				},
			},
		},
	})
}

func registerSelectionPages() {
	lago.RegistryPage.Register("programs.ProgramSelectionTable", &components.Modal{
		UID: "program-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Program]{
				Page:  components.Page{Key: "programs.ProgramSelectionTableBody"},
				UID:   "program-selection-table",
				Title: "Select Program",
				Data:  getters.Key[components.ObjectList[Program]]("programs"),
				RowAttr: getters.RowAttrSelect("ProgramID", getters.Key[uint]("$row.ID"), ProgramDisplayLabel(
					getters.Key[string]("$row.Name"),
					getters.Key[string]("$row.University"),
				)),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "programs.ProgramSelectionFilter"}},
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
						Label: "University",
						Name:  "University",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.University")},
						},
					},
					{
						Label: "Program type",
						Name:  "ProgramType",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.ProgramType")},
						},
					},
				},
			},
		},
	})
}

func registerProgramMediaMultiSelectPages() {
	lago.RegistryPage.Register("programs.ProgramMediaMultiSelectionTable", &components.Modal{
		UID: "program-media-multi-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[ProgramMedia]{
				Page:  components.Page{Key: "programs.ProgramMediaMultiSelectionTableBody"},
				UID:   "program-media-multi-selection-table",
				Title: "Select languages",
				Data:  getters.Key[components.ObjectList[ProgramMedia]]("program_media"),
				RowAttr: getters.RowAttrSelectMulti(
					getters.IfOrElse(
						getters.Key[string]("$get.target_input"),
						getters.Static("ProgramMedia"),
					),
					getters.Key[uint]("$row.ID"),
					getters.Key[string]("$row.Language"),
				),
				Columns: []components.TableColumn{
					{
						Label: "Language",
						Name:  "Language",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Language")},
						},
					},
				},
			},
		},
	})
}
