package p_seer_deepsearch

import (
	"net/http"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func init() {
	deepSearchListPatchers := views.QueryPatchers[DeepSearch]{
		{Key: "seer_deepsearch.list.not_deleted", Value: deepSearchActiveOnlyPatcher{}},
		{Key: "seer_deepsearch.list.order", Value: views.QueryPatcherOrderBy[DeepSearch]{Order: "id DESC"}},
	}

	deepSearchDetailPatchers := views.QueryPatchers[DeepSearch]{
		{Key: "seer_deepsearch.detail_preload_logs", Value: views.QueryPatcherPreload[DeepSearch]{
			Fields: []string{"Logs"},
			PreloadBuilder: func(_ views.View, _ *http.Request, pb gorm.PreloadBuilder) error {
				pb.Order(`"created_at" DESC`)
				return nil
			},
		}},
	}

	lago.RegistryView.Register("seer_deepsearch.HomeView",
		lago.GetPageView("seer_deepsearch.DeepSearchHome").
			WithLayer("users.auth", p_users.AuthenticationLayer{}))

	lago.RegistryView.Register("seer_deepsearch.HistoryView",
		lago.GetPageView("seer_deepsearch.HistoryTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_deepsearch.deepsearch.list", views.LayerList[DeepSearch]{
				Key:           getters.Static("deepSearches"),
				QueryPatchers: deepSearchListPatchers,
			}))

	lago.RegistryView.Register("seer_deepsearch.StartView",
		lago.GetPageView("seer_deepsearch.StartBlank").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_deepsearch.start_method", deepSearchStartRejectGetLayer{}).
			WithLayer("seer_deepsearch.start_post", deepSearchStartPostLayer{}))

	lago.RegistryView.Register("seer_deepsearch.DetailView",
		lago.GetPageView("seer_deepsearch.DeepSearchDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_deepsearch.deepsearch.detail", views.LayerDetail[DeepSearch]{
				Key:           getters.Static("deepSearch"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: deepSearchDetailPatchers,
			}))

	lago.RegistryView.Register("seer_deepsearch.StopView",
		lago.GetPageView("seer_deepsearch.StartBlank").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_deepsearch.stop_method", deepSearchStartRejectGetLayer{}).
			WithLayer("seer_deepsearch.stop_post", deepSearchStopPostLayer{}))

	lago.RegistryView.Register("seer_deepsearch.RestartView",
		lago.GetPageView("seer_deepsearch.StartBlank").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_deepsearch.restart_method", deepSearchStartRejectGetLayer{}).
			WithLayer("seer_deepsearch.restart_post", deepSearchRestartPostLayer{}))
}
