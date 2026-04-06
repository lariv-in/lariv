package p_totschool_tally

import (
	"context"
	"fmt"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

// tallySessionEnvironmentDefault returns the TotSchoolSession id for the current quarter.
func registerTallyUserDetailPatch() {
	// Patch the users.UserDetail page using InsertChildAfter to append
	// a session environment input that allows changing the active session.
	lago.RegistryPage.Patch("users.UserDetail", func(page components.PageInterface) components.PageInterface {
		if scaffold, ok := page.(*components.ShellScaffold); ok {
			// Ensure ApexCharts is loaded in the page head for StatLineChart.
			// NOTE: We originally attempted to inject ApexCharts into ExtraHead here,
			// but ContainerHTML requires a gomponents.Node signature and there is
			// currently no HTML wrapper in this package. To keep linting clean, we
			// skip injecting ExtraHead for now; StatLineChart assumes ApexCharts is
			// available globally (e.g. via the base layout).

			// Insert an environment input after the main user detail content
			// so that the session variable can be changed from this page.
			components.InsertChildAfter(scaffold,
				"users.UserDetailContent",
				func(*components.Detail[p_users.User]) components.ContainerColumn {
					return components.ContainerColumn{
						Children: []components.PageInterface{
							&components.Environment[uint]{
								Label:   "Session",
								Key:     getters.Static("session"),
								Options: SessionsListGetter,
								Default: tallySessionEnvironmentDefault,
							},
							TallySessionEntries{
								Page: components.Page{
									Key: "tally.UserSessionTallies",
								},
								UserGetter:    getters.Key[p_users.User]("user"),
								SessionGetter: CurrentEnvironmentSessionGetter,
							},
							StatLineChart{
								Page: components.Page{
									Key: "tally.UserSessionTalliesChart",
								},
								TalliesGetter: func(ctx context.Context) ([]Tally, error) {
									db, ok := ctx.Value("$db").(*gorm.DB)
									if !ok || db == nil {
										return nil, fmt.Errorf("StatLineChart: missing $db in context")
									}
									user, ok := ctx.Value("user").(p_users.User)
									if !ok {
										return nil, fmt.Errorf("StatLineChart: missing user in context")
									}
									session, err := CurrentEnvironmentSessionGetter(ctx)
									if err != nil {
										return nil, err
									}
									var tallies []Tally
									if err := db.
										Where("user_id = ? AND date >= ? AND date <= ?", user.ID, session.Start, session.End).
										Order("date ASC").
										Find(&tallies).Error; err != nil {
										return nil, err
									}
									return tallies, nil
								},
								Keys: []string{
									"Visits",
									"Appointments",
									"Leads",
									"Presentations",
									"Demos",
									"Letters",
									"FollowUps",
									"Proposals",
									"Policies",
									"Premium",
								},
							},
						},
					}
				},
			)

			return scaffold
		}
		panic("Base page for users.UserDetail was not ShellScaffold")
	})
}
