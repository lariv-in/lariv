package p_nirmancampus_students

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

// studentFormUserPickURL opens the scoped user picker; on edit, allow_user_id keeps the linked user visible.
func registerStudentUserPickPages() {
	lago.RegistryPage.Register("students.UserPickFilter", &components.FormComponent[p_users.User]{
		Attr: getters.FormBoostedGet(lago.RoutePath("students.UserPickRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Name:   "allow_user_id",
				Hidden: true,
				Getter: getters.Key[string]("$get.allow_user_id"),
			},
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
			&components.InputText{Label: "Email", Name: "Email", Getter: getters.Key[string]("$get.Email")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("students.UserPickTable", &components.Modal{
		UID: "student-user-pick-modal",
		Children: []components.PageInterface{
			&components.DataTable[p_users.User]{
				UID:     "student-user-pick-table",
				Title:   "Select User",
				Data:    getters.Key[components.ObjectList[p_users.User]]("users"),
				RowAttr: getters.RowAttrSelect("UserID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "students.UserPickFilter"}},
					&components.ButtonModalForm{
						Name:        getters.Static("users.UserCreateForm"),
						Url:         lago.RoutePath("users.CreateRoute", nil),
						FormPostURL: lago.RoutePath("users.CreateRoute", nil),
						ModalUID:    "user-create-modal",
						Icon:        "plus",
						Classes:     "btn-square btn-outline btn-sm",
						Attr:        getters.ModalRefreshList(getters.Static(""), getters.Static("#student-user-pick-table")),
					},
				},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
					{Label: "Email", Name: "Email", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Email")},
					}},
					{Label: "Phone", Name: "Phone", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Phone")},
					}},
				},
			},
		},
	})
}
