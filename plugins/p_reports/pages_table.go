package p_reports

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("reports.ReportDefinitionTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "reports.ReportDefinitionMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[ReportDefinition]{
				Page: components.Page{Key: "reports.ReportDefinitionTableBody"}, UID: "report-definition-table", Classes: "w-full",
				Data:    getters.Key[components.ObjectList[ReportDefinition]]("report_definitions"),
				Actions: []components.PageInterface{&components.TableButtonCreate{Link: lago.RoutePath("reports.CreateRoute", nil)}},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("reports.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
				},
			},
		},
	})
}
