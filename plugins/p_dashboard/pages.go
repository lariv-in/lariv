package p_dashboard

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/lago"
	pcomps "github.com/lariv-in/lago/plugins/p_dashboard/components"
)

func init() {
	components.RegistryTopbar.Register("dashboard.appsPageButton", components.ButtonLink{
		Icon:    "squares-2x2",
		Link:    lago.GetterRoutePath("dashboard.AppsPage", nil),
		Classes: "btn-sm btn-square btn-neutral",
	})
	components.RegistryTopbar.Register("dashboard.themeButton", pcomps.ThemeButton{
		Classes: "btn-sm btn-square btn-outline",
	})
	components.RegistryTopbar.Register("dashboard.userDropdown", pcomps.UserDropdown{})

	lago.RegistryPage.Register("dashboard.AppsPage", components.ShellTopbarScaffold{
		Children: []components.PageInterface{
			components.LayoutSimple{
				Children: []components.PageInterface{
					pcomps.AppsGrid{},
				},
			},
		},
	})
}
