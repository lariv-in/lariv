package p_seer_intel

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("seer_intel.IntelTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_intel.IntelMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Intel]{
				Page:    components.Page{Key: "seer_intel.IntelTableBody"},
				UID:     "seer-intel-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Intel]]("intels"),
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_intel.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Title")},
						},
					},
					{
						Label: "Kind",
						Name:  "Kind",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Kind")},
						},
					},
					{
						Label: "Datetime",
						Name:  "Datetime",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Datetime")},
						},
					},
				},
			},
		},
	})
}
