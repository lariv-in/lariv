package p_users

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("users.UserFilter", &components.FormComponent[User]{
		Attr: getters.FormBoostedGet(lago.RoutePath("users.ListRoute", nil)),

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
	})

	lago.RegistryPage.Register("users.UserSelectionFilter", &components.FormComponent[User]{
		Attr: getters.FormBoostedGet(lago.RoutePath("users.SelectRoute", nil)),

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
	})

	lago.RegistryPage.Register("users.RoleSelectionFilter", &components.FormComponent[Role]{
		Attr: getters.FormBoostedGet(lago.RoutePath("users.SelectRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})
}
