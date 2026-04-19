package p_seer_reddit

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func init() {
	sourcePatchers := views.QueryPatchers[RedditSource]{
		{Key: "seer_reddit.source.order", Value: views.QueryPatcherOrderBy[RedditSource]{Order: "id DESC"}},
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
			}))

	lago.RegistryView.Register("seer_reddit.RedditSourceCreateView",
		lago.GetPageView("seer_reddit.RedditSourceCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.create", views.LayerCreate[RedditSource]{
				SuccessURL: lago.RoutePath("seer_reddit.RedditSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_reddit.source.create_validate", Value: redditSourceCreateValidate{}},
				},
			}))

	lago.RegistryView.Register("seer_reddit.RedditSourceUpdateView",
		lago.GetPageView("seer_reddit.RedditSourceUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.detail", views.LayerDetail[RedditSource]{
				Key:          getters.Static("redditSource"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_reddit.reddit_source.update", views.LayerUpdate[RedditSource]{
				Key: getters.Static("redditSource"),
				SuccessURL: lago.RoutePath("seer_reddit.RedditSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditSource.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_reddit.source.create_validate", Value: redditSourceCreateValidate{}},
				},
			}))

	lago.RegistryView.Register("seer_reddit.RedditSourceDeleteView",
		lago.GetPageView("seer_reddit.RedditSourceDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.delete_detail", views.LayerDetail[RedditSource]{
				Key:          getters.Static("redditSource"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_reddit.reddit_source.delete", views.LayerDelete[RedditSource]{
				Key:        getters.Static("redditSource"),
				SuccessURL: lago.RoutePath("seer_reddit.DefaultRoute", nil),
			}))

	postPatchers := views.QueryPatchers[RedditPost]{
		{Key: "seer_reddit.post.not_deleted", Value: redditPostActiveOnlyPatcher{}},
		{Key: "seer_reddit.post.order", Value: views.QueryPatcherOrderBy[RedditPost]{Order: "id DESC"}},
	}

	postDetailPatchers := views.QueryPatchers[RedditPost]{
		{Key: "seer_reddit.post_detail.not_deleted", Value: redditPostActiveOnlyPatcher{}},
	}

	lago.RegistryView.Register("seer_reddit.RedditPostListView",
		lago.GetPageView("seer_reddit.RedditPostTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_post.list", views.LayerList[RedditPost]{
				Key:           getters.Static("redditPosts"),
				QueryPatchers: postPatchers,
			}))

	lago.RegistryView.Register("seer_reddit.RedditPostListBySourceView",
		lago.GetPageView("seer_reddit.RedditPostTableBySource").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.detail_by_source_id", views.LayerDetail[RedditSource]{
				Key:          getters.Static("redditSource"),
				PathParamKey: getters.Static("source_id"),
			}).
			WithLayer("seer_reddit.reddit_post.list_by_source", views.LayerList[RedditPost]{
				Key:           getters.Static("redditPosts"),
				QueryPatchers: redditPostListQueryPatchersForSource(),
			}))

	lago.RegistryView.Register("seer_reddit.RedditPostListBulkAddIntelView",
		lago.GetPageView("seer_reddit.RedditPostTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_post.list", views.LayerList[RedditPost]{
				Key:           getters.Static("redditPosts"),
				QueryPatchers: postPatchers,
			}).
			WithLayer("seer_reddit.reddit_posts_bulk_intel_list", redditPostsBulkAddIntelLayer{
				redirectRouteName: "seer_reddit.RedditPostListRoute",
			}))

	lago.RegistryView.Register("seer_reddit.RedditPostListBySourceBulkAddIntelView",
		lago.GetPageView("seer_reddit.RedditPostTableBySource").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.detail_by_source_id", views.LayerDetail[RedditSource]{
				Key:          getters.Static("redditSource"),
				PathParamKey: getters.Static("source_id"),
			}).
			WithLayer("seer_reddit.reddit_post.list_by_source", views.LayerList[RedditPost]{
				Key:           getters.Static("redditPosts"),
				QueryPatchers: redditPostListQueryPatchersForSource(),
			}).
			WithLayer("seer_reddit.reddit_posts_bulk_intel_by_source", redditPostsBulkAddIntelLayer{
				redirectRouteName: "seer_reddit.RedditPostListBySourceRoute",
				sourceIDPathParam: "source_id",
			}))

	lago.RegistryView.Register("seer_reddit.RedditSourceFetchPostsView",
		lago.GetPageView("seer_reddit.RedditPostTableBySource").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.fetch_posts", redditSourceFetchPostsActionLayer{}))

	lago.RegistryView.Register("seer_reddit.RedditPostDetailView",
		lago.GetPageView("seer_reddit.RedditPostDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_post.detail", views.LayerDetail[RedditPost]{
				Key:           getters.Static("redditPost"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: postDetailPatchers,
			}))

	lago.RegistryView.Register("seer_reddit.RedditPostAddIntelView",
		lago.GetPageView("seer_reddit.RedditPostDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_post.add_intel_detail", views.LayerDetail[RedditPost]{
				Key:           getters.Static("redditPost"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: postDetailPatchers,
			}).
			WithLayer("seer_reddit.reddit_post.add_intel", redditPostAddIntelLayer{}))

	lago.RegistryView.Register("seer_reddit.RedditPostSoftDeleteView",
		lago.GetPageView("seer_reddit.RedditPostDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_post.delete_detail", views.LayerDetail[RedditPost]{
				Key:           getters.Static("redditPost"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: postDetailPatchers,
			}).
			WithLayer("seer_reddit.reddit_post.soft_delete", redditPostSoftDeleteLayer{}))

	runnerPatchers := views.QueryPatchers[RedditRunner]{
		{Key: "seer_reddit.runner.order", Value: views.QueryPatcherOrderBy[RedditRunner]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("seer_reddit.RedditRunnerListView",
		lago.GetPageView("seer_reddit.RedditRunnerTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.list", views.LayerList[RedditRunner]{
				Key:           getters.Static("redditRunners"),
				QueryPatchers: runnerPatchers,
			}))

	lago.RegistryView.Register("seer_reddit.RedditRunnerSelectView",
		lago.GetPageView("seer_reddit.RedditRunnerSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.select_list", views.LayerList[RedditRunner]{
				Key:           getters.Static("redditRunners"),
				QueryPatchers: runnerPatchers,
			}))

	lago.RegistryView.Register("seer_reddit.RedditRunnerDetailView",
		lago.GetPageView("seer_reddit.RedditRunnerDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.detail", views.LayerDetail[RedditRunner]{
				Key:          getters.Static("redditRunner"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("seer_reddit.RedditRunnerCreateView",
		lago.GetPageView("seer_reddit.RedditRunnerCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.create", views.LayerCreate[RedditRunner]{
				SuccessURL: lago.RoutePath("seer_reddit.RedditRunnerDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_reddit.runner.validate", Value: redditRunnerValidate{}},
				},
			}))

	lago.RegistryView.Register("seer_reddit.RedditRunnerUpdateView",
		lago.GetPageView("seer_reddit.RedditRunnerUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.detail", views.LayerDetail[RedditRunner]{
				Key:          getters.Static("redditRunner"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_reddit.reddit_runner.update", views.LayerUpdate[RedditRunner]{
				Key: getters.Static("redditRunner"),
				SuccessURL: lago.RoutePath("seer_reddit.RedditRunnerDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditRunner.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_reddit.runner.validate", Value: redditRunnerValidate{}},
				},
			}))

	lago.RegistryView.Register("seer_reddit.RedditRunnerDeleteView",
		lago.GetPageView("seer_reddit.RedditRunnerDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.delete_detail", views.LayerDetail[RedditRunner]{
				Key:          getters.Static("redditRunner"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_reddit.reddit_runner.delete", views.LayerDelete[RedditRunner]{
				Key:        getters.Static("redditRunner"),
				SuccessURL: lago.RoutePath("seer_reddit.RedditRunnerListRoute", nil),
			}))
}

func redditPostListQueryPatchersForSource() views.QueryPatchers[RedditPost] {
	return views.QueryPatchers[RedditPost]{
		{Key: "seer_reddit.post.not_deleted", Value: redditPostActiveOnlyPatcher{}},
		{Key: "seer_reddit.post.order", Value: views.QueryPatcherOrderBy[RedditPost]{Order: "id DESC"}},
		{Key: "seer_reddit.post.for_current_source", Value: redditPostsForCurrentSourcePatcher{}},
	}
}

// redditPostActiveOnlyPatcher scopes queries to rows with [gorm.Model.DeletedAt] unset (non–soft-deleted).
type redditPostActiveOnlyPatcher struct{}

func (redditPostActiveOnlyPatcher) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[RedditPost]) gorm.ChainInterface[RedditPost] {
	return q.Where("deleted_at IS NULL")
}

type redditSourceCreateValidate struct{}

func (redditSourceCreateValidate) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	subRaw, ok := formData["Subreddits"]
	if !ok {
		formErrors["Subreddits"] = errors.New("add at least one subreddit")
		return formData, formErrors
	}
	var b []byte
	switch v := subRaw.(type) {
	case datatypes.JSON:
		b = []byte(v)
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		formErrors["Subreddits"] = errors.New("invalid subreddits value")
		return formData, formErrors
	}
	var subs []string
	if err := json.Unmarshal(b, &subs); err != nil {
		formErrors["Subreddits"] = err
		return formData, formErrors
	}
	n := 0
	for _, s := range subs {
		if strings.TrimSpace(s) != "" {
			n++
		}
	}
	if n == 0 {
		formErrors["Subreddits"] = errors.New("add at least one subreddit")
	}
	rid, ok := redditRunnerIDFromFormMap(formData)
	if !ok || rid == 0 {
		formErrors["RedditRunnerID"] = errors.New("choose a worker")
	}
	return formData, formErrors
}

func redditRunnerIDFromFormMap(formData map[string]any) (uint, bool) {
	v, ok := formData["RedditRunnerID"]
	if !ok {
		return 0, false
	}
	rid, ok := v.(uint)
	if !ok {
		return 0, false
	}
	return rid, true
}

type redditRunnerValidate struct{}

func (redditRunnerValidate) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	name, _ := formData["Name"].(string)
	if strings.TrimSpace(name) == "" {
		formErrors["Name"] = errors.New("name is required")
	}
	durRaw, ok := formData["Duration"]
	if !ok {
		formErrors["Duration"] = errors.New("duration is required")
		return formData, formErrors
	}
	d, ok := durRaw.(*time.Duration)
	if !ok {
		formErrors["Duration"] = errors.New("invalid duration")
		return formData, formErrors
	}
	if d == nil || *d <= 0 {
		formErrors["Duration"] = errors.New("duration must be positive")
	}
	return formData, formErrors
}
