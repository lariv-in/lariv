package p_nirmancampus_studentapplications

import (
	"context"
	"fmt"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
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

func applicationCreateUrlGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		role, err := getters.Key[string]("$role")(ctx)
		if err != nil {
			return "", err
		}
		if role == "superuser" || role == "admin" || role == roleNameUnassigned {
			return lago.RoutePath("studentapplications.CreateRoute", nil)(ctx)
		}
		return "", fmt.Errorf("you do not have permission to do this action")
	}
}

func registerTablePages() {
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
					&components.TableButtonCreate{Link: applicationCreateUrlGetter()},
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
							&components.FieldText{Getter: getters.Key[string]("$row.Program.Name")},
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
