package p_lacerate

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	patchers := views.QueryPatchers[Report]{
		{Key: "lacerate.reports.order_id", Value: views.QueryPatcherOrderBy[Report]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("lacerate.ReportListView",
		lago.GetPageView("lacerate.ReportsTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reports.list", views.LayerList[Report]{
				Key:           getters.Static("reports"),
				QueryPatchers: patchers,
			}))

	lago.RegistryView.Register("lacerate.ReportDetailView",
		lago.GetPageView("lacerate.ReportDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reports.detail", views.LayerDetail[Report]{
				Key:          getters.Static("report"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("lacerate.ReportCreateView",
		lago.GetPageView("lacerate.ReportCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reports.create", views.LayerCreate[Report]{
				SuccessURL: lago.RoutePath("lacerate.ReportDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("lacerate.ReportUpdateView",
		lago.GetPageView("lacerate.ReportUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reports.update_detail", views.LayerDetail[Report]{
				Key:          getters.Static("report"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.reports.update", views.LayerUpdate[Report]{
				Key: getters.Static("report"),
				SuccessURL: lago.RoutePath("lacerate.ReportDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("report.ID")),
				}),
			}))

	lago.RegistryView.Register("lacerate.ReportDeleteView",
		lago.GetPageView("lacerate.ReportDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reports.delete_detail", views.LayerDetail[Report]{
				Key:          getters.Static("report"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.reports.delete", views.LayerDelete[Report]{
				Key:        getters.Static("report"),
				SuccessURL: lago.RoutePath("lacerate.ReportListRoute", nil),
			}))
}
