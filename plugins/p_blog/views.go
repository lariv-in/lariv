package p_blog

import (
	"net/http"
	"strings"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_users"
	"github.com/lariv-in/lariv/registry"
	"github.com/lariv-in/lariv/views"
)

type AuthorFormPatcher struct{}

func (p AuthorFormPatcher) Patch(view views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if user, ok := p_users.UserFromContextOptional(r.Context()); ok {
		if val, exists := formData["CreatedByID"]; !exists || val == nil || val == uint(0) || val == float64(0) || val == "" {
			formData["CreatedByID"] = user.ID
			delete(formErrors, "CreatedByID")
		}
	}
	return formData, formErrors
}

type SlugFormPatcher struct{}

func (p SlugFormPatcher) Patch(view views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	slugVal, _ := formData["Slug"].(string)
	if strings.TrimSpace(slugVal) == "" {
		if titleVal, ok := formData["Title"].(string); ok && strings.TrimSpace(titleVal) != "" {
			formData["Slug"] = getters.TitleToFormSlug(titleVal)
		} else {
			formData["Slug"] = "blog"
		}
	} else {
		formData["Slug"] = getters.TitleToFormSlug(slugVal)
	}
	delete(formErrors, "Slug")
	return formData, formErrors
}

func pluginViews() lariv.PluginFeatures[*views.View] {
	return lariv.PluginFeatures[*views.View]{
		Entries: []registry.Pair[string, *views.View]{
			// Blog Views
			{
				Key: "p_blog.BlogListView",
				Value: lariv.GetPageView("p_blog.BlogListPage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_blog.list", views.LayerList[Blog]{
						Key: getters.Static("blogs"),
						QueryPatchers: views.QueryPatchers[Blog]{
							{Key: "p_blog.blogs.preload", Value: views.QueryPatcherPreload[Blog]{Fields: []string{"CreatedBy", "Tags"}}},
						},
					}),
			},
			{
				Key: "p_blog.BlogDetailView",
				Value: lariv.GetPageView("p_blog.BlogDetailPage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_blog.detail", views.LayerDetail[Blog]{
						Key:          getters.Static("blog"),
						PathParamKey: getters.Static("id"),
						QueryPatchers: views.QueryPatchers[Blog]{
							{Key: "p_blog.blog.preload", Value: views.QueryPatcherPreload[Blog]{Fields: []string{"CreatedBy", "Tags"}}},
						},
					}),
			},
			{
				Key: "p_blog.BlogCreateView",
				Value: lariv.GetPageView("p_blog.BlogCreatePage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_blog.create", views.LayerCreate[Blog]{
						SuccessURL: lariv.RoutePath("p_blog.BlogListRoute", nil),
						FormPatchers: views.FormPatchers{
							{Key: "p_blog.author_patcher", Value: AuthorFormPatcher{}},
							{Key: "p_blog.slug_patcher", Value: SlugFormPatcher{}},
						},
					}),
			},
			{
				Key: "p_blog.BlogUpdateView",
				Value: lariv.GetPageView("p_blog.BlogUpdatePage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_blog.detail", views.LayerDetail[Blog]{
						Key:          getters.Static("blog"),
						PathParamKey: getters.Static("id"),
						QueryPatchers: views.QueryPatchers[Blog]{
							{Key: "p_blog.blog.preload", Value: views.QueryPatcherPreload[Blog]{Fields: []string{"CreatedBy", "Tags"}}},
						},
					}).
					WithLayer("p_blog.update", views.LayerUpdate[Blog]{
						Key: getters.Static("blog"),
						SuccessURL: lariv.RoutePath("p_blog.BlogDetailRoute", map[string]getters.Getter[any]{
							"id": getters.Any(getters.Key[uint]("blog.ID")),
						}),
						FormPatchers: views.FormPatchers{
							{Key: "p_blog.slug_patcher", Value: SlugFormPatcher{}},
						},
					}),
			},
			{
				Key: "p_blog.BlogDeleteView",
				Value: lariv.GetPageView("p_blog.BlogDeleteForm").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_blog.detail", views.LayerDetail[Blog]{
						Key:          getters.Static("blog"),
						PathParamKey: getters.Static("id"),
					}).
					WithLayer("p_blog.delete", views.LayerDelete[Blog]{
						Key:        getters.Static("blog"),
						SuccessURL: lariv.RoutePath("p_blog.BlogListRoute", nil),
					}),
			},

			// BlogTag Views
			{
				Key: "p_blog.TagListView",
				Value: lariv.GetPageView("p_blog.TagListPage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_blog.tags.list", views.LayerList[BlogTag]{
						Key: getters.Static("tags"),
					}),
			},
			{
				Key: "p_blog.TagDetailView",
				Value: lariv.GetPageView("p_blog.TagDetailPage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_blog.tags.detail", views.LayerDetail[BlogTag]{
						Key:          getters.Static("tag"),
						PathParamKey: getters.Static("id"),
						QueryPatchers: views.QueryPatchers[BlogTag]{
							{Key: "p_blog.tag.preload", Value: views.QueryPatcherPreload[BlogTag]{Fields: []string{"Blogs"}}},
						},
					}),
			},
			{
				Key: "p_blog.TagCreateView",
				Value: lariv.GetPageView("p_blog.TagCreatePage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_blog.tags.create", views.LayerCreate[BlogTag]{
						SuccessURL: lariv.RoutePath("p_blog.TagListRoute", nil),
					}),
			},
			{
				Key: "p_blog.TagUpdateView",
				Value: lariv.GetPageView("p_blog.TagUpdatePage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_blog.tags.detail", views.LayerDetail[BlogTag]{
						Key:          getters.Static("tag"),
						PathParamKey: getters.Static("id"),
					}).
					WithLayer("p_blog.tags.update", views.LayerUpdate[BlogTag]{
						Key: getters.Static("tag"),
						SuccessURL: lariv.RoutePath("p_blog.TagDetailRoute", map[string]getters.Getter[any]{
							"id": getters.Any(getters.Key[uint]("tag.ID")),
						}),
					}),
			},
			{
				Key: "p_blog.TagDeleteView",
				Value: lariv.GetPageView("p_blog.TagDeleteForm").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_blog.tags.detail", views.LayerDetail[BlogTag]{
						Key:          getters.Static("tag"),
						PathParamKey: getters.Static("id"),
					}).
					WithLayer("p_blog.tags.delete", views.LayerDelete[BlogTag]{
						Key:        getters.Static("tag"),
						SuccessURL: lariv.RoutePath("p_blog.TagListRoute", nil),
					}),
			},
			{
				Key: "p_blog.TagSelectView",
				Value: lariv.GetPageView("p_blog.TagSelectionTable").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_blog.tags.select", views.LayerList[BlogTag]{
						Key: getters.Static("tags"),
					}),
			},
		},
	}
}
