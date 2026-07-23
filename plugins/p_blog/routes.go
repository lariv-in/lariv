package p_blog

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

const (
	TagURL     = "/blog/tags/"
	BlogIDPath = AppURL + "p/{id}/"
	TagIDPath  = TagURL + "{id}/"
)

func pluginRoutes() lariv.PluginFeatures[lariv.Route] {
	return lariv.PluginFeatures[lariv.Route]{
		Entries: []registry.Pair[string, lariv.Route]{
			// Blog Routes
			{Key: "p_blog.BlogListRoute", Value: lariv.Route{
				Path:    AppURL,
				Handler: lariv.NewDynamicView("p_blog.BlogListView"),
			}},
			{Key: "p_blog.BlogCreateRoute", Value: lariv.Route{
				Path:    AppURL + "create/",
				Handler: lariv.NewDynamicView("p_blog.BlogCreateView"),
			}},
			{Key: "p_blog.BlogDetailRoute", Value: lariv.Route{
				Path:    BlogIDPath,
				Handler: lariv.NewDynamicView("p_blog.BlogDetailView"),
			}},
			{Key: "p_blog.BlogUpdateRoute", Value: lariv.Route{
				Path:    BlogIDPath + "edit/",
				Handler: lariv.NewDynamicView("p_blog.BlogUpdateView"),
			}},
			{Key: "p_blog.BlogDeleteRoute", Value: lariv.Route{
				Path:    BlogIDPath + "delete/",
				Handler: lariv.NewDynamicView("p_blog.BlogDeleteView"),
			}},

			// BlogTag Routes
			{Key: "p_blog.TagListRoute", Value: lariv.Route{
				Path:    TagURL,
				Handler: lariv.NewDynamicView("p_blog.TagListView"),
			}},
			{Key: "p_blog.TagCreateRoute", Value: lariv.Route{
				Path:    TagURL + "create/",
				Handler: lariv.NewDynamicView("p_blog.TagCreateView"),
			}},
			{Key: "p_blog.TagDetailRoute", Value: lariv.Route{
				Path:    TagIDPath,
				Handler: lariv.NewDynamicView("p_blog.TagDetailView"),
			}},
			{Key: "p_blog.TagUpdateRoute", Value: lariv.Route{
				Path:    TagIDPath + "edit/",
				Handler: lariv.NewDynamicView("p_blog.TagUpdateView"),
			}},
			{Key: "p_blog.TagDeleteRoute", Value: lariv.Route{
				Path:    TagIDPath + "delete/",
				Handler: lariv.NewDynamicView("p_blog.TagDeleteView"),
			}},
			{Key: "p_blog.TagSelectRoute", Value: lariv.Route{
				Path:    TagURL + "select/",
				Handler: lariv.NewDynamicView("p_blog.TagSelectView"),
			}},
		},
	}
}
