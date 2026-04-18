package p_seer_reddit

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	sourcePatchers := views.QueryPatchers[RedditSource]{
		{Key: "seer_reddit.source.order", Value: views.QueryPatcherOrderBy[RedditSource]{Order: "id DESC"}},
		{Key: "seer_reddit.source.preload", Value: views.QueryPatcherPreload[RedditSource]{Fields: []string{"Runner"}}},
	}

	lago.RegistryView.Register("seer_reddit.RedditSourceListView",
		lago.GetPageView("seer_reddit.RedditSourceTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.list", views.LayerList[RedditSource]{
				Key:           getters.Static("redditSources"),
				QueryPatchers: sourcePatchers,
			}))

	lago.RegistryView.Register("seer_reddit.RedditSourceDetailView",
		lago.GetPageView("seer_reddit.RedditSourceDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.detail", views.LayerDetail[RedditSource]{
				Key:          getters.Static("redditSource"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[RedditSource]{
					{Key: "seer_reddit.source.preload", Value: views.QueryPatcherPreload[RedditSource]{Fields: []string{"Runner"}}},
				},
			}))

	postPatchers := views.QueryPatchers[RedditPost]{
		{Key: "seer_reddit.post.order", Value: views.QueryPatcherOrderBy[RedditPost]{Order: "id DESC"}},
		{Key: "seer_reddit.post.preload", Value: views.QueryPatcherPreload[RedditPost]{Fields: []string{"RedditSource"}}},
	}

	lago.RegistryView.Register("seer_reddit.RedditPostListView",
		lago.GetPageView("seer_reddit.RedditPostTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_post.list", views.LayerList[RedditPost]{
				Key:           getters.Static("redditPosts"),
				QueryPatchers: postPatchers,
			}))

	lago.RegistryView.Register("seer_reddit.RedditPostDetailView",
		lago.GetPageView("seer_reddit.RedditPostDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_post.detail", views.LayerDetail[RedditPost]{
				Key:          getters.Static("redditPost"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[RedditPost]{
					{Key: "seer_reddit.post.preload", Value: views.QueryPatcherPreload[RedditPost]{Fields: []string{"RedditSource"}}},
				},
			}))
}
