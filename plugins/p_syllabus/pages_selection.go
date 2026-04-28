package p_syllabus

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerSelectionPages() {
	lago.RegistryPage.Register("syllabus.SyllabusTopicMultiSelectionFilter", &components.FormComponent[SyllabusTopic]{
		Attr: getters.FormBoostedGet(lago.RoutePath("syllabus.MultiSelectRoute", nil)),
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Title", Name: "Title", Getter: getters.Key[string]("$get.Title")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})
}
