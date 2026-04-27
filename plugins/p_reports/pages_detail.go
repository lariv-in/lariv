package p_reports

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

func registerDetailPages() {
	lago.RegistryPage.Register("reports.ReportDefinitionDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "reports.ReportDefinitionDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[ReportDefinition]{
				Getter: getters.Key[ReportDefinition]("report_definition"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "reports.ReportDefinitionDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.LabelInline{Title: "Frequency", Children: []components.PageInterface{
								&components.FieldText{Getter: registry.PairValueFromKey(getters.Key[string]("$in.Frequency"), ReportFrequencyChoices)},
							}},
							&components.LabelInline{Title: "Report at", Children: []components.PageInterface{
								&components.FieldDatetime{Getter: getters.Deref(getters.Key[*time.Time]("$in.ReportAt"))},
							}},
							&components.LabelInline{Title: "Notes", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Notes")},
							}},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("reports.ReportDefinitionDeleteForm", &components.Modal{
		UID: "report-definition-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm deletion", Message: "Delete this report definition?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
