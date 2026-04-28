package p_reports

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

func reportDefinitionFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "reports.ReportDefinitionFormFields"},
		Children: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.Key[string]("$in.Name")},
			&components.InputTextarea{Label: "Notes", Name: "Notes", Rows: 4, Getter: getters.Key[string]("$in.Notes")},
			&components.InputSelect[string]{Label: "Frequency", Name: "Frequency", Required: true, Choices: getters.Static(ReportFrequencyChoices), Getter: registry.PairFromGetter(getters.Key[string]("$in.Frequency"), ReportFrequencyChoices)},
			&components.InputDatetime{Label: "Report at (optional)", Name: "ReportAt", Required: false, Getter: getters.Deref(getters.Key[*time.Time]("$in.ReportAt"))},
		},
	}
}

func registerFormPages() {
	dn := getters.Static("reports.ReportDefinitionDeleteForm")
	lago.RegistryPage.Register("reports.ReportDefinitionCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "reports.ReportDefinitionMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("reports.ReportDefinitionCreateForm"),
				ActionURL: lago.RoutePath("reports.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[ReportDefinition]{
						Attr:           getters.FormBubbling(getters.Static("reports.ReportDefinitionCreateForm")),
						Title:          "Create report definition",
						ChildrenInput:  []components.PageInterface{reportDefinitionFormFields()},
						ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save"}},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("reports.ReportDefinitionUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "reports.ReportDefinitionDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("reports.ReportDefinitionUpdateForm"),
				ActionURL: lago.RoutePath("reports.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("report_definition.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[ReportDefinition]{
						Getter:        getters.Key[ReportDefinition]("report_definition"),
						Attr:          getters.FormBubbling(getters.Static("reports.ReportDefinitionUpdateForm")),
						Title:         "Edit",
						ChildrenInput: []components.PageInterface{reportDefinitionFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label: "Delete", Icon: "trash", Name: dn,
										Url:         lago.RoutePath("reports.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("report_definition.ID"))}),
										FormPostURL: lago.RoutePath("reports.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("report_definition.ID"))}),
										ModalUID:    "report-definition-delete-modal", Classes: "btn-error",
									},
								},
							},
						},
					},
				},
			},
		},
	})
}
