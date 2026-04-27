package p_teachers

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerSelectionPages() {
	lago.RegistryPage.Register("teachers.TeacherSelectionFilter", &components.FormComponent[Teacher]{
		Attr: getters.FormBoostedGet(lago.RoutePath("teachers.SelectRoute", nil)),
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.Key[string]("$get.Code")},
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
	lago.RegistryPage.Register("teachers.TeacherMultiSelectionFilter", &components.FormComponent[Teacher]{
		Attr: getters.FormBoostedGet(lago.RoutePath("teachers.MultiSelectRoute", nil)),
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.Key[string]("$get.Code")},
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
}
