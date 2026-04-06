package p_users

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("users.UserTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[User]{
				UID:     "user-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[User]]("users"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "users.UserFilter"}},
					&components.ButtonModalForm{
						Name:        getters.Static("users.UserCreateForm"),
						Url:         lago.RoutePath("users.CreateRoute", nil),
						FormPostURL: lago.RoutePath("users.CreateRoute", nil),
						ModalUID:    "user-create-modal",
						Icon:        "plus",
						Classes:     "btn-square btn-outline btn-sm",
					},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("users.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
					{Label: "Email", Name: "Email", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Email")},
					}},
					{Label: "Phone", Name: "Phone", Children: []components.PageInterface{
						&components.FieldPhone{Getter: getters.Key[string]("$row.Phone")},
					}},
				},
			},
		},
	})
}

// --- Detail & Delete ---
