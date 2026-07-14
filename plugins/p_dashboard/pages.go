package p_dashboard

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/components"
	pcomps "github.com/lariv-in/lago/plugins/p_dashboard/components"
	"github.com/lariv-in/lago/registry"
)

func init() {
	components.RegistryTopbar.Register("dashboard.appsPageButton", components.ButtonLink{
		Icon:    "squares-2x2",
		Link:    lago.RoutePath("dashboard.AppsPage", nil),
		Classes: "btn-sm btn-square btn-neutral",
	})
	components.RegistryTopbar.Register("dashboard.themeButton", pcomps.ThemeButton{
		Classes: "btn-sm btn-square btn-outline",
	})
	components.RegistryTopbar.Register("dashboard.userDropdown", pcomps.UserDropdown{})
}

func pluginPages() lago.PluginFeatures[components.PageInterface] {
	return lago.PluginFeatures[components.PageInterface]{
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
