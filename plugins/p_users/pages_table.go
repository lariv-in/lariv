package p_users

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func pageEntriesTables() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		{Key: "p_users.UserTable", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_users.UserMenu"},
			},
			Children: []components.PageInterface{
				&components.DataTable[User]{
					UID:     "user-table",
					Classes: "w-full",
					Data:    getters.Key[components.ObjectList[User]]("users"),
					Actions: []components.PageInterface{
						&components.TableButtonFilter{Child: lariv.DynamicPage{Name: "p_users.UserFilter"}},
						&components.ButtonModalForm{
							Name:        getters.Static("p_users.UserCreateForm"),
							Url:         lariv.RoutePath("p_users.CreateRoute", nil),
							FormPostURL: lariv.RoutePath("p_users.CreateRoute", nil),
							ModalUID:    "user-create-modal",
							Icon:        "plus",
							Classes:     "btn-square btn-outline btn-sm",
						},
					},
					RowAttr: getters.RowAttrNavigate(lariv.RoutePath("p_users.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
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
		}},
	}
}
