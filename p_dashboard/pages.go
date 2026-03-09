package p_dashboard

import (
	"github.com/lariv-in/components"
	"github.com/lariv-in/lago"
	pcomps "github.com/lariv-in/p_dashboard/components"
)

func init() {
	lago.RegistryPage.Register("dashboard.AppsPage", components.LayoutTopbarScaffold{
		Children: []components.PageInterface{
			components.LayoutSimple{
				Children: []components.PageInterface{
					pcomps.AppsGrid{},
				},
			},
		},
	})
}
