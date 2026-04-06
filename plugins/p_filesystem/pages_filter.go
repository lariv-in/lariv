package p_filesystem

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerFilters() {
	lago.RegistryPage.Register("filesystem.VNodeFilter", &components.FormComponent[VNode]{
		Attr: getters.FormBoostedGet(listOrBrowseRoute("filesystem.ListRoute", "filesystem.BrowseRoute")),

		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply Filters"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("filesystem.ParentSelectionFilter", &components.FormComponent[VNode]{
		Attr: getters.FormBoostedGet(withSelectionTarget(listOrBrowseRoute("filesystem.SelectRoute", "filesystem.SelectChildRoute"))),

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

	lago.RegistryPage.Register("filesystem.DestinationSelectionFilter", &components.FormComponent[VNode]{
		Attr: getters.FormBoostedGet(withSelectionTarget(listOrBrowseRoute("filesystem.MoveSelectRoute", "filesystem.MoveSelectChildRoute"))),

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
