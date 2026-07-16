package p_dashboard

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	pcomps "github.com/lariv-in/lariv/plugins/p_dashboard/components"
	"github.com/lariv-in/lariv/registry"
)

func init() {
	components.RegistryTopbar.Register("dashboard.appsPageButton", components.ButtonLink{
		Icon:    "squares-2x2",
		Link:    lariv.RoutePath("dashboard.AppsPage", nil),
		Classes: "btn-sm btn-square btn-neutral",
	})
	components.RegistryTopbar.Register("dashboard.themeButton", pcomps.ThemeButton{
		Classes: "btn-sm btn-square btn-outline",
	})
	components.RegistryTopbar.Register("dashboard.userDropdown", pcomps.UserDropdown{})
}

func pluginPages() lariv.PluginFeatures[components.PageInterface] {
	return lariv.PluginFeatures[components.PageInterface]{
		Entries: []registry.Pair[string, components.PageInterface]{
			{Key: "dashboard.HomeRedirectStub", Value: &components.ContainerColumn{
				Page:     components.Page{Key: "dashboard.HomeRedirectStub"},
				Children: []components.PageInterface{},
			}},
			{Key: "dashboard.AppsPage", Value: &components.ShellTopbarScaffold{
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
			}},
		},
	}
}
