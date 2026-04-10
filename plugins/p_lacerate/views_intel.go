package p_lacerate

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

func init() {
	intelListPatchers := views.QueryPatchers[Intel]{
		{Key: "lacerate.intel.preload_source", Value: views.QueryPatcherPreload[Intel]{Field: "Source"}},
		{Key: "lacerate.intel.preload_preview", Value: views.QueryPatcherPreload[Intel]{Field: "PreviewImage"}},
		{Key: "lacerate.intel.order_id", Value: views.QueryPatcherOrderBy[Intel]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("lacerate.IntelListView",
		lago.GetPageView("lacerate.IntelTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.intel.list", views.LayerList[Intel]{
				Key:           getters.Static("intels"),
				QueryPatchers: intelListPatchers,
			}))

	lago.RegistryView.Register("lacerate.IntelDetailView",
		lago.GetPageView("lacerate.IntelDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.intel.detail", views.LayerDetail[Intel]{
				Key:          getters.Static("intel"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Intel]{
					{Key: "lacerate.intel.detail_preload", Value: views.QueryPatcherPreload[Intel]{
						Field: "Source",
					}},
					{Key: "lacerate.intel.detail_preload_preview", Value: views.QueryPatcherPreload[Intel]{
						Field: "PreviewImage",
					}},
				},
			}))

	lago.RegistryView.Register("lacerate.IntelCreateView",
		lago.GetPageView("lacerate.IntelCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.intel.create", views.LayerCreate[Intel]{
				SuccessURL: lago.RoutePath("lacerate.IntelDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("lacerate.IntelUpdateView",
		lago.GetPageView("lacerate.IntelUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.intel.update_detail", views.LayerDetail[Intel]{
				Key:          getters.Static("intel"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Intel]{
					{Key: "lacerate.intel.update_preload_src", Value: views.QueryPatcherPreload[Intel]{Field: "Source"}},
					{Key: "lacerate.intel.update_preload_preview", Value: views.QueryPatcherPreload[Intel]{Field: "PreviewImage"}},
				},
			}).
			WithLayer("lacerate.intel.update", views.LayerUpdate[Intel]{
				Key: getters.Static("intel"),
				SuccessURL: lago.RoutePath("lacerate.IntelDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("intel.ID")),
				}),
			}))

	lago.RegistryView.Register("lacerate.IntelDeleteView",
		lago.GetPageView("lacerate.IntelDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.intel.delete_detail", views.LayerDetail[Intel]{
				Key:          getters.Static("intel"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Intel]{
					{Key: "lacerate.intel.delete_preload", Value: views.QueryPatcherPreload[Intel]{Field: "Source"}},
				},
			}).
			WithLayer("lacerate.intel.delete", views.LayerDelete[Intel]{
				Key:        getters.Static("intel"),
				SuccessURL: lago.RoutePath("lacerate.IntelListRoute", nil),
			}))

	lago.RegistryView.Register("lacerate.SourceSelectView",
		lago.GetPageView("lacerate.SourceSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.sources.select", views.LayerList[Source]{
				Key: getters.Static("sources"),
				QueryPatchers: views.QueryPatchers[Source]{
					registry.Pair[string, views.QueryPatcher[Source]]{
						Key:   "lacerate.sources.order_id",
						Value: views.QueryPatcherOrderBy[Source]{Order: "id DESC"},
					},
				},
			}))
}
