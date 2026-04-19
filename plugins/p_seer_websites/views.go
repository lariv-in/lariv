package p_seer_websites

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	websitePatchers := views.QueryPatchers[Website]{
		{Key: "seer_websites.website.not_deleted", Value: websiteActiveOnlyPatcher{}},
		{Key: "seer_websites.website.order", Value: views.QueryPatcherOrderBy[Website]{Order: "id DESC"}},
	}

	websiteDetailPatchers := views.QueryPatchers[Website]{
		{Key: "seer_websites.website_detail.not_deleted", Value: websiteActiveOnlyPatcher{}},
	}

	lago.RegistryView.Register("seer_websites.WebsiteListView",
		lago.GetPageView("seer_websites.WebsiteTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website.list", views.LayerList[Website]{
				Key:           getters.Static("websites"),
				QueryPatchers: websitePatchers,
			}))

	lago.RegistryView.Register("seer_websites.WebsiteAddView",
		lago.GetPageView("seer_websites.WebsiteAddForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website.create", views.LayerCreate[Website]{
				SuccessURL: lago.RoutePath("seer_websites.WebsiteDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_websites.website.scrape", Value: websiteScrapeFormPatcher{}},
				},
			}))

	lago.RegistryView.Register("seer_websites.WebsiteDetailView",
		lago.GetPageView("seer_websites.WebsiteDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website.detail", views.LayerDetail[Website]{
				Key:           getters.Static("website"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: websiteDetailPatchers,
			}))

	lago.RegistryView.Register("seer_websites.WebsiteAddIntelView",
		lago.GetPageView("seer_websites.WebsiteDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website.add_intel_detail", views.LayerDetail[Website]{
				Key:           getters.Static("website"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: websiteDetailPatchers,
			}).
			WithLayer("seer_websites.website.add_intel", websiteAddIntelLayer{}))

	lago.RegistryView.Register("seer_websites.WebsiteSoftDeleteView",
		lago.GetPageView("seer_websites.WebsiteDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website.delete_detail", views.LayerDetail[Website]{
				Key:           getters.Static("website"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: websiteDetailPatchers,
			}).
			WithLayer("seer_websites.website.soft_delete", websiteSoftDeleteLayer{}))
}
