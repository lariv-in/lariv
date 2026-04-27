package p_programs

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerSelectionPages() {
	lago.RegistryPage.Register("programs.ProgramSelectionFilter", &components.FormComponent[Program]{
		Attr: getters.FormBoostedGet(lago.RoutePath("programs.SelectRoute", nil)),
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.Key[string]("$get.Code")},
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})
	lago.RegistryPage.Register("programs.ProgramMultiSelectionFilter", &components.FormComponent[Program]{
		Attr: getters.FormBoostedGet(lago.RoutePath("programs.MultiSelectRoute", nil)),
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Code", Name: "Code", Getter: getters.Key[string]("$get.Code")},
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
