package p_seer_reddit

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerRedditSourcePages()
	registerRedditPostPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("seer_reddit.RedditMenu", &components.SidebarMenu{
		Title: getters.Static("Reddit"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Sources"),
				Url:   lago.RoutePath("seer_reddit.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Posts"),
				Url:   lago.RoutePath("seer_reddit.RedditPostListRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("seer_reddit.RedditSourceDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Reddit source #%d", getters.Any(getters.Key[uint]("redditSource.ID"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Sources"),
			Url:   lago.RoutePath("seer_reddit.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("seer_reddit.RedditSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditSource.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("seer_reddit.RedditPostDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Post #%d", getters.Any(getters.Key[uint]("redditPost.ID"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Posts"),
			Url:   lago.RoutePath("seer_reddit.RedditPostListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("seer_reddit.RedditPostDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditPost.ID")),
				}),
			},
		},
	})
}
