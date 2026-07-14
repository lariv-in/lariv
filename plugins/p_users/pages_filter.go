package p_users

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
)

func pageEntriesFilters() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		{Key: "p_users.UserFilter", Value: &components.FormComponent[User]{
			Attr: getters.FormBoostedGet(lago.RoutePath("p_users.ListRoute", nil)),

			ChildrenInput: []components.PageInterface{
				&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
				&components.InputText{Label: "Email", Name: "Email", Getter: getters.Key[string]("$get.Email")},
				&components.InputPhone{Label: "Phone", Name: "Phone", Getter: getters.Key[string]("$get.Phone")},
			},
			ChildrenAction: []components.PageInterface{
				components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply Filters"},
					&components.ButtonClear{Label: "Clear"},
				}},
			},
		}},
		{Key: "p_users.UserSelectionFilter", Value: &components.FormComponent[User]{
			Attr: getters.FormBoostedGet(lago.RoutePath("p_users.SelectRoute", nil)),

			ChildrenInput: []components.PageInterface{
				&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
				&components.InputText{Label: "Email", Name: "Email", Getter: getters.Key[string]("$get.Email")},
			},
			ChildrenAction: []components.PageInterface{
				&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply"},
					&components.ButtonClear{Label: "Clear"},
				}},
			},
		}},
		{Key: "p_users.RoleSelectionFilter", Value: &components.FormComponent[Role]{
			Attr: getters.FormBoostedGet(lago.RoutePath("p_users.SelectRoute", nil)),

			ChildrenInput: []components.PageInterface{
				&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
			},
			ChildrenAction: []components.PageInterface{
				&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply"},
					&components.ButtonClear{Label: "Clear"},
				}},
			},
		}},
	}
}
