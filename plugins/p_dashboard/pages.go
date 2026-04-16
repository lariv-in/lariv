package p_dashboard

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/lago"
	pcomps "github.com/lariv-in/lago/plugins/p_dashboard/components"
)

func init() {
	// Empty shell for root redirect view (layers redirect; page never renders).
	lago.RegistryPage.Register("dashboard.HomeRedirectStub", &components.ContainerColumn{
		Page:     components.Page{Key: "dashboard.HomeRedirectStub"},
		Children: []components.PageInterface{},
	})

	components.RegistryTopbar.Register("dashboard.appsPageButton", components.ButtonLink{
		Icon:    "squares-2x2",
		Link:    lago.RoutePath("dashboard.AppsPage", nil),
		Classes: "btn-sm btn-square btn-neutral",
	})
	components.RegistryTopbar.Register("dashboard.themeButton", pcomps.ThemeButton{
		Classes: "btn-sm btn-square btn-outline",
	})
	components.RegistryTopbar.Register("dashboard.userDropdown", pcomps.UserDropdown{})

	lago.RegistryPage.Register("dashboard.AppsPage", &components.ShellTopbarScaffold{
		Children: []components.PageInterface{
			&components.LayoutSimple{
				Page: components.Page{Key: "dashboard.AppsPageLayout"},
				Children: []components.PageInterface{
					&pcomps.AppsGrid{
						Page: components.Page{Key: "dashboard.AppsGrid"},
					},
				},
			},
		},
	})
}
