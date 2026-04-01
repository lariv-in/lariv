package p_nirmancampus_studentapplications

import (
	"context"
	"log"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerDashboardAppsPagePatch()
}

func getterRoleIsUnassigned() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		role, err := getters.Key[string]("$role")(ctx)
		if err != nil {
			return false, nil
		}
		return role == roleNameUnassigned, nil
	}
}

func getterRoleIsNotUnassigned() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		role, err := getters.Key[string]("$role")(ctx)
		if err != nil {
			return true, nil
		}
		return role != roleNameUnassigned, nil
	}
}

func registerDashboardAppsPagePatch() {
	lago.RegistryPage.Patch("dashboard.AppsPage", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellTopbarScaffold)
		if !ok {
			log.Panic("dashboard.AppsPage was not *components.ShellTopbarScaffold")
		}
		components.ReplaceChild(scaffold, "dashboard.AppsPageLayout", func(layout *components.LayoutSimple) *components.LayoutSimple {
			if len(layout.Children) != 1 {
				log.Panic("dashboard.AppsPageLayout: expected exactly one child (AppsGrid)")
			}
			appsGrid := layout.Children[0]
			layout.Children = []components.PageInterface{
				&components.ShowIf{
					Page:   components.Page{Key: "studentapplications.DashboardUnassignedActions"},
					Getter: getterRoleIsUnassigned(),
					Children: []components.PageInterface{
						components.ContainerColumn{
							Page:    components.Page{Key: "studentapplications.DashboardUnassignedColumn"},
							Classes: "gap-4",
							Children: []components.PageInterface{
								&components.FieldText{
									Page:    components.Page{Key: "studentapplications.DashboardUnassignedHello"},
									Getter:  getters.Format("Hello %s", getters.Any(getters.Key[string]("$user.Name"))),
									Classes: "text-3xl font-bold mb-8",
								},
								&components.ContainerRow{
									Classes: "flex-wrap gap-3",
									Children: []components.PageInterface{
										&components.ButtonLink{
											Label:   "Create application",
											Link:    lago.RoutePath("studentapplications.CreateRoute", nil),
											Icon:    "plus",
											Classes: "btn-primary",
										},
										&components.ButtonLink{
											Label:   "View your applications",
											Link:    lago.RoutePath("studentapplications.DefaultRoute", nil),
											Icon:    "document-text",
											Classes: "btn-outline",
										},
									},
								},
							},
						},
					},
				},
				&components.ShowIf{
					Page:   components.Page{Key: "studentapplications.DashboardAppsGrid"},
					Getter: getterRoleIsNotUnassigned(),
					Children: []components.PageInterface{
						appsGrid,
					},
				},
			}
			return layout
		})
		return scaffold
	})
}
