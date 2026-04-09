package p_nirmancampus_students

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerFilterPages() {
	lago.RegistryPage.Register("students.StudentFilter", &components.FormComponent[Student]{
		Attr: getters.FormBoostedGet(lago.RoutePath("students.DefaultRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Enrollment No / Control ID",
				Name:   "StudentNo",
				Getter: getters.Key[string]("$get.StudentNo"),
			},
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.Key[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Email",
				Name:   "Email",
				Getter: getters.Key[string]("$get.Email"),
			},
			&components.InputText{
				Label:  "Phone",
				Name:   "Phone",
				Getter: getters.Key[string]("$get.Phone"),
			},
			&components.InputText{
				Label:  "Mother's Name",
				Name:   "MotherName",
				Getter: getters.Key[string]("$get.MotherName"),
			},
			&components.InputText{
				Label:  "Father's Name",
				Name:   "FatherName",
				Getter: getters.Key[string]("$get.FatherName"),
			},
			&components.InputText{
				Label:  "Category",
				Name:   "Category",
				Getter: getters.Key[string]("$get.Category"),
			},
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

	lago.RegistryPage.Register("students.StudentSelectionFilter", &components.FormComponent[Student]{
		Attr: getters.FormBoostedGet(lago.RoutePath("students.SelectRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.Key[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Enrollment No / Control ID",
				Name:   "StudentNo",
				Getter: getters.Key[string]("$get.StudentNo"),
			},
			&components.InputText{
				Label:  "Phone",
				Name:   "Phone",
				Getter: getters.Key[string]("$get.Phone"),
			},
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
					&components.TableButtonFilter{
						Child: lago.DynamicPage{Name: "students.StudentFilter"},
						Page:  components.Page{Roles: []string{"admin", "superuser"}},
					},
					&components.TableButtonCreate{
						Link: lago.RoutePath("students.CreateRoute", nil),
						Page: components.Page{Roles: []string{"admin", "superuser"}},
					},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Key[string]("$row.Name"),
							},
						},
					},
					{
						Label: "Enrollment No / Control ID",
						Name:  "StudentNo",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Key[string]("$row.StudentNo"),
							},
						},
					},
					{
						Label: "Email",
						Name:  "Email",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Key[string]("$row.Email"),
							},
						},
					},
					{
						Label: "Phone",
						Name:  "Phone",
						Children: []components.PageInterface{
							&components.FieldPhone{
								Getter: getters.Key[string]("$row.Phone"),
							},
						},
					},
				},
			},
		},
	})
}

func registerSelectionPages() {
	lago.RegistryPage.Register("students.StudentSelectionTable", &components.Modal{
		UID: "student-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Student]{
				Page:    components.Page{Key: "students.StudentSelectionTableBody"},
				UID:     "student-selection-table",
				Title:   "Select Student",
				Data:    getters.Key[components.ObjectList[Student]]("students"),
				RowAttr: getters.RowAttrSelect("StudentID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "students.StudentSelectionFilter"}},
				},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Key[string]("$row.Name"),
							},
						},
					},
					{
						Label: "Enrollment No / Control ID",
						Name:  "StudentNo",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Key[string]("$row.StudentNo"),
							},
						},
					},
					{
						Label: "Phone",
						Name:  "Phone",
						Children: []components.PageInterface{
							&components.FieldPhone{
								Getter: getters.Key[string]("$row.Phone"),
							},
						},
					},
				},
			},
		},
	})
}
