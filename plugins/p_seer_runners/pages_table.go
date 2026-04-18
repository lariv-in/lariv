package p_seer_runners

import (
	"context"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func formatRunnerDuration(d time.Duration) string {
	if d == 0 {
		return "—"
	}
	return d.String()
}

func registerTablePages() {
	lago.RegistryPage.Register("seer_runners.RunnerTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_runners.RunnerMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Runner]{
				Page:    components.Page{Key: "seer_runners.RunnerTableBody"},
				UID:     "seer-runners-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Runner]]("runners"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("seer_runners.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_runners.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "ID",
						Name:  "ID",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.ID")))},
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
						Label: "Duration",
						Name:  "Duration",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Map(getters.Key[time.Duration]("$row.Duration"), func(_ context.Context, d time.Duration) (string, error) {
									return formatRunnerDuration(d), nil
								}),
							},
						},
					},
				},
			},
		},
	})
}
