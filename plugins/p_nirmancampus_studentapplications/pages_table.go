package p_nirmancampus_studentapplications

import (
	"fmt"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
)

func registerFilterPages() {
	lago.RegistryPage.Register("studentapplications.ApplicationFilter", &components.FormComponent[StudentApplication]{
		Attr: getters.FormBoostedGet(lago.RoutePath("studentapplications.DefaultRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Email",
				Name:   "Email",
				Getter: getters.Key[string]("$get.Email"),
			},
			&components.InputText{
				Label:  "Student name",
				Name:   "StudentName",
				Getter: getters.Key[string]("$get.StudentName"),
			},
			&components.InputText{
				Label:  "Mobile",
				Name:   "Mobile",
				Getter: getters.Key[string]("$get.Mobile"),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply filters"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})
}

func registerTablePages() {
	create := lago.RoutePath("studentapplications.CreateRoute", nil)
	studentApplicationTableCreateLink := getters.Match(getters.Key[string]("$role"), map[string]getters.Getter[string]{
		"superuser":        create,
		"admin":            create,
		roleNameUnassigned: create,
	}, getters.Static(fmt.Errorf("you do not have permission to do this action")))
	lago.RegistryPage.Register("studentapplications.ApplicationTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "studentapplications.ApplicationMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[StudentApplication]{
				Page:    components.Page{Key: "studentapplications.ApplicationTableBody"},
				UID:     "student-application-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[StudentApplication]]("studentapplications"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "studentapplications.ApplicationFilter"}},
					&components.TableButtonCreate{Link: studentApplicationTableCreateLink},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{
						Label: "Email",
						Name:  "Email",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Email")},
						},
					},
					{
						Label: "Program",
						Name:  "Program.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: p_nirmancampus_programs.ProgramDisplayLabel(
								getters.Key[string]("$row.Program.Name"),
								getters.Key[string]("$row.Program.University"),
							)},
						},
					},
					{
						Label: "Student",
						Name:  "StudentName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.StudentName")},
						},
					},
					{
						Label: "Mobile",
						Name:  "Mobile",
						Children: []components.PageInterface{
							&components.FieldPhone{Getter: getters.Key[string]("$row.Mobile")},
						},
					},
				},
			},
		},
	})
}
