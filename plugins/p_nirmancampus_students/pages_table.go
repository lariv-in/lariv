package p_nirmancampus_students

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

// studentFormUserPickURL opens the scoped user picker; on edit, allow_user_id keeps the linked user visible.
func registerFilterPages() {
	lago.RegistryPage.Register("students.StudentFilter", &components.FormComponent[Student]{
		Attr: getters.FormBoostedGet(lago.RoutePath("students.DefaultRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Student Number",
				Name:   "StudentNo",
				Getter: getters.Key[string]("$get.StudentNo"),
			},
			&components.InputText{
				Label:  "Name",
				Name:   "User.Name",
				Getter: getters.Key[string]("$get.User.Name"),
			},
			&components.InputText{
				Label:  "Email",
				Name:   "User.Email",
				Getter: getters.Key[string]("$get.User.Email"),
			},
			&components.InputText{
				Label:  "Phone",
				Name:   "User.Phone",
				Getter: getters.Key[string]("$get.User.Phone"),
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
				Name:   "User.Name",
				Getter: getters.Key[string]("$get.User.Name"),
			},
			&components.InputText{
				Label:  "Student No",
				Name:   "StudentNo",
				Getter: getters.Key[string]("$get.StudentNo"),
			},
			&components.InputText{
				Label:  "Phone",
				Name:   "User.Phone",
				Getter: getters.Key[string]("$get.User.Phone"),
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
						Name:  "User.Name",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.ForeignKey[p_users.User, uint, string](
									getters.Key[uint]("$row.UserID"),
									"Name",
								),
							},
						},
					},
					{
						Label: "Student No",
						Name:  "StudentNo",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Key[string]("$row.StudentNo"),
							},
						},
					},
					{
						Label: "Email",
						Name:  "User.Email",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.ForeignKey[p_users.User, uint, string](
									getters.Key[uint]("$row.UserID"),
									"Email",
								),
							},
						},
					},
					{
						Label: "Phone",
						Name:  "User.Phone",
						Children: []components.PageInterface{
							&components.FieldPhone{
								Getter: getters.ForeignKey[p_users.User, uint, string](
									getters.Key[uint]("$row.UserID"),
									"Phone",
								),
							},
						},
					},
					{
						Label: "Mother's Name",
						Name:  "MotherName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.MotherName")},
						},
					},
					{
						Label: "Father's Name",
						Name:  "FatherName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.FatherName")},
						},
					},
					{
						Label: "Category",
						Name:  "Category",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Category")},
						},
					},
					{
						Label: "Address",
						Name:  "Address",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Address")},
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
				RowAttr: getters.RowAttrSelect("StudentID", getters.Key[uint]("$row.ID"), getters.ForeignKey[Student, uint, string](getters.Key[uint]("$row.ID"), "StudentNo")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "students.StudentSelectionFilter"}},
				},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "User.Name",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.ForeignKey[p_users.User, uint, string](
									getters.Key[uint]("$row.UserID"),
									"Name",
								),
							},
						},
					},
					{
						Label: "Student No",
						Name:  "StudentNo",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Key[string]("$row.StudentNo"),
							},
						},
					},
					{
						Label: "Phone",
						Name:  "User.Phone",
						Children: []components.PageInterface{
							&components.FieldPhone{
								Getter: getters.ForeignKey[p_users.User, uint, string](
									getters.Key[uint]("$row.UserID"),
									"Phone",
								),
							},
						},
					},
					{
						Label: "Mother's Name",
						Name:  "MotherName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.MotherName")},
						},
					},
					{
						Label: "Father's Name",
						Name:  "FatherName",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.FatherName")},
						},
					},
					{
						Label: "Category",
						Name:  "Category",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Category")},
						},
					},
				},
			},
		},
	})
}
