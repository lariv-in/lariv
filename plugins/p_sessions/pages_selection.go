package p_sessions

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerSelectionPages() {
	lago.RegistryPage.Register("sessions.ClassSessionSelectionFilter", &components.FormComponent[ClassSession]{
		Attr: getters.FormBoostedGet(lago.RoutePath("sessions.SelectRoute", nil)),
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Title", Name: "Title", Getter: getters.Key[string]("$get.Title")},
			&components.InputText{Label: "Room", Name: "Room", Getter: getters.Key[string]("$get.Room")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})
}
