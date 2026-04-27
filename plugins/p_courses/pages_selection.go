package p_courses

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerSelectionPages() {
	lago.RegistryPage.Register("courses.CourseSelectionFilter", &components.FormComponent[Course]{
		Attr: getters.FormBoostedGet(lago.RoutePath("courses.SelectRoute", nil)),
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
