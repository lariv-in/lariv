package p_users

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerSelectionPages() {
	lago.RegistryPage.Register("users.UserSelectionTable", &components.Modal{
		UID: "user-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[User]{
				UID:     "user-selection-table",
				Title:   "Select User",
				Data:    getters.Key[components.ObjectList[User]]("users"),
				RowAttr: getters.RowAttrSelect("UserID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "users.UserSelectionFilter"}},
					&components.ButtonModalForm{
						Name:        getters.Static("users.UserCreateForm"),
						Url:         lago.RoutePath("users.CreateRoute", nil),
						FormPostURL: lago.RoutePath("users.CreateRoute", nil),
						ModalUID:    "user-create-modal",
						Icon:        "plus",
						Classes:     "btn-square btn-outline btn-sm",
						Attr:        getters.ModalRefreshList(getters.Static(""), getters.Static("#user-selection-table")),
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

	lago.RegistryPage.Register("users.RoleSelectionTable", &components.Modal{
		UID: "role-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Role]{
				UID:     "role-selection-table",
				Title:   "Select Role",
				Data:    getters.Key[components.ObjectList[Role]]("roles"),
				RowAttr: getters.RowAttrSelect("RoleID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "users.RoleSelectionFilter"}},
					&components.ButtonModalForm{
						Name:        getters.Static("users.RoleCreateForm"),
						Url:         lago.RoutePath("users.RoleCreateRoute", nil),
						FormPostURL: lago.RoutePath("users.RoleCreateRoute", nil),
						ModalUID:    "role-create-modal",
						Icon:        "plus",
						Classes:     "btn-square btn-outline btn-sm",
						Attr:        getters.ModalRefreshList(getters.Static(""), getters.Static("#role-selection-table")),
					},
				},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
				},
			},
		},
	})
}

// --- Role CRUD Pages ---
