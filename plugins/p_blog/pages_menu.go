package p_blog

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func pageEntriesMenus() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		// Main Navigation Menu
		{Key: "p_blog.BlogListMenu", Value: &components.SidebarMenu{
			Title: getters.Static("Blog Admin"),
			Back: &components.SidebarMenuItem{
				Title: getters.Static("Back to All Apps"),
				Url:   lariv.RoutePath("dashboard.AppsPage", nil),
			},
			Children: []components.PageInterface{
				&components.SidebarMenuItem{
					Title: getters.Static("All Articles"),
					Url:   lariv.RoutePath("p_blog.BlogListRoute", nil),
				},
				&components.SidebarMenuItem{
					Title: getters.Static("Blog Tags"),
					Url:   lariv.RoutePath("p_blog.TagListRoute", nil),
				},
			},
		}},

		// Blog Article Detail Menu
		{Key: "p_blog.BlogDetailMenu", Value: &components.SidebarMenu{
			Title: getters.Format("Blog: %s", getters.Any(getters.Key[string]("blog.Title"))),
			Back: &components.SidebarMenuItem{
				Title: getters.Static("Back to All Articles"),
				Url:   lariv.RoutePath("p_blog.BlogListRoute", nil),
			},
			Children: []components.PageInterface{
				&components.SidebarMenuItem{
					Title: getters.Static("Article Details"),
					Url: lariv.RoutePath("p_blog.BlogDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("blog.ID")),
					}),
				},
				&components.SidebarMenuItem{
					Title: getters.Static("Edit Article"),
					Url: lariv.RoutePath("p_blog.BlogUpdateRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("blog.ID")),
					}),
				},
			},
		}},

		// Blog Tag Detail Menu
		{Key: "p_blog.TagDetailMenu", Value: &components.SidebarMenu{
			Title: getters.Format("Tag: %s", getters.Any(getters.Key[string]("tag.Name"))),
			Back: &components.SidebarMenuItem{
				Title: getters.Static("Back to All Tags"),
				Url:   lariv.RoutePath("p_blog.TagListRoute", nil),
			},
			Children: []components.PageInterface{
				&components.SidebarMenuItem{
					Title: getters.Static("Tag Details"),
					Url: lariv.RoutePath("p_blog.TagDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("tag.ID")),
					}),
				},
				&components.SidebarMenuItem{
					Title: getters.Static("Edit Tag"),
					Url: lariv.RoutePath("p_blog.TagUpdateRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("tag.ID")),
					}),
				},
			},
		}},
	}
}
