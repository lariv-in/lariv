package p_students

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerSelectionPages() {
	lago.RegistryPage.Register("students.StudentSelectionFilter", &components.FormComponent[Student]{
		Attr: getters.FormBoostedGet(lago.RoutePath("students.SelectRoute", nil)),
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Student No", Name: "StudentNo", Getter: getters.Key[string]("$get.StudentNo")},
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
	lago.RegistryPage.Register("students.StudentMultiSelectionFilter", &components.FormComponent[Student]{
		Attr: getters.FormBoostedGet(lago.RoutePath("students.MultiSelectRoute", nil)),
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Student No", Name: "StudentNo", Getter: getters.Key[string]("$get.StudentNo")},
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
