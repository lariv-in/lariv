package p_lacerate

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	patchers := views.QueryPatchers[Lookup]{
		{Key: "lacerate.lookups.order_id", Value: views.QueryPatcherOrderBy[Lookup]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("lacerate.LookupListView",
		lago.GetPageView("lacerate.LookupsTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.lookups.list", views.LayerList[Lookup]{
				Key:           getters.Static("lookups"),
				QueryPatchers: patchers,
			}))

	lago.RegistryView.Register("lacerate.LookupDetailView",
		lago.GetPageView("lacerate.LookupDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.lookups.detail", views.LayerDetail[Lookup]{
				Key:          getters.Static("lookup"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.lookups.detail_logs", lookupDetailLogsLayer{}))

	lago.RegistryView.Register("lacerate.LookupCreateView",
		lago.GetPageView("lacerate.LookupCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.lookups.create", views.LayerCreate[Lookup]{
				SuccessURL: lago.RoutePath("lacerate.LookupDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("lacerate.LookupUpdateView",
		lago.GetPageView("lacerate.LookupUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.lookups.update_detail", views.LayerDetail[Lookup]{
				Key:          getters.Static("lookup"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.lookups.update", views.LayerUpdate[Lookup]{
				Key: getters.Static("lookup"),
				SuccessURL: lago.RoutePath("lacerate.LookupDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("lookup.ID")),
				}),
			}))

	lago.RegistryView.Register("lacerate.LookupDeleteView",
		lago.GetPageView("lacerate.LookupDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.lookups.delete_detail", views.LayerDetail[Lookup]{
				Key:          getters.Static("lookup"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.lookups.delete", views.LayerDelete[Lookup]{
				Key:        getters.Static("lookup"),
				SuccessURL: lago.RoutePath("lacerate.LookupListRoute", nil),
			}))
}
