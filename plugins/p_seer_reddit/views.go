package p_seer_reddit

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func init() {
	sourcePatchers := views.QueryPatchers[RedditSource]{
		{Key: "seer_reddit.source.order", Value: views.QueryPatcherOrderBy[RedditSource]{Order: "id DESC"}},
	}
	sourceDetailPatchers := views.QueryPatchers[RedditSource]{
		{Key: "seer_reddit.source.preload_runner", Value: views.QueryPatcherPreload[RedditSource]{Fields: []string{"RedditRunner"}}},
	}
	sourceUnsetPatchers := views.QueryPatchers[RedditSource]{
		{Key: "seer_reddit.source.unset_runner", Value: redditSourceUnsetRunnerPatcher{}},
		{Key: "seer_reddit.source.order", Value: views.QueryPatcherOrderBy[RedditSource]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("seer_reddit.RedditSourceListView",
		lago.GetPageView("seer_reddit.RedditSourceTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.list", views.LayerList[RedditSource]{
				Key:           getters.Static("redditSources"),
				QueryPatchers: sourcePatchers,
			}))

	lago.RegistryView.Register("seer_reddit.RedditSourceUnsetSelectView",
		lago.GetPageView("seer_reddit.RedditSourceUnsetSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.unset_select_list", views.LayerList[RedditSource]{
				Key:           getters.Static("redditSources"),
				QueryPatchers: sourceUnsetPatchers,
			}))

	lago.RegistryView.Register("seer_reddit.RedditSourceDetailView",
		lago.GetPageView("seer_reddit.RedditSourceDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.detail", views.LayerDetail[RedditSource]{
				Key:           getters.Static("redditSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: sourceDetailPatchers,
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
				Key:           getters.Static("redditSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: sourceDetailPatchers,
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
				Key:           getters.Static("redditSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: sourceDetailPatchers,
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
				Key:           getters.Static("redditSource"),
				PathParamKey:  getters.Static("source_id"),
				QueryPatchers: sourceDetailPatchers,
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
				Key:           getters.Static("redditSource"),
				PathParamKey:  getters.Static("source_id"),
				QueryPatchers: sourceDetailPatchers,
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
			WithLayer("seer_reddit.reddit_runner.enrich_source_ids", redditRunnerEnrichSourceIDsLayer{}).
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

type redditSourceUnsetRunnerPatcher struct{}

func (redditSourceUnsetRunnerPatcher) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[RedditSource]) gorm.ChainInterface[RedditSource] {
	return q.Where("reddit_runner_id IS NULL")
}

type redditSourceCreateValidate struct{}

func (redditSourceCreateValidate) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	p, err := RedditSourceCreateParamsFromFormMap(formData)
	if err != nil {
		formErrors["Subreddits"] = err
		return formData, formErrors
	}
	for k, v := range ValidateRedditSourceCreate(p) {
		formErrors[k] = v
	}
	return formData, formErrors
}

func redditRunnerIDFromFormMap(formData map[string]any) (*uint, bool) {
	v, ok := formData["RedditRunnerID"]
	if !ok {
		return nil, false
	}
	rid, ok := v.(uint)
	if !ok {
		return nil, false
	}
	if rid == 0 {
		return nil, true
	}
	return new(rid), true
}

type redditRunnerValidate struct{}

func (redditRunnerValidate) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if formErrors == nil {
		formErrors = map[string]error{}
	}
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
	formData, formErrors = redditRunnerSourceIDsValidateAndFlatten(formData, formErrors)
	formErrors = validateRedditRunnerSourceIDs(r, formData, formErrors)
	return formData, formErrors
}

func redditRunnerSourceIDsValidateAndFlatten(formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	raw, ok := formData["RedditSourceIDs"]
	if !ok {
		return formData, formErrors
	}
	assoc, ok := raw.(components.AssociationIDs)
	if !ok {
		formErrors["RedditSourceIDs"] = errors.New("invalid Reddit sources")
		delete(formData, "RedditSourceIDs")
		return formData, formErrors
	}
	formData["RedditSourceIDs"] = assoc.IDs
	return formData, formErrors
}

func validateRedditRunnerSourceIDs(r *http.Request, formData map[string]any, formErrors map[string]error) map[string]error {
	ids, _ := formData["RedditSourceIDs"].([]uint)
	if len(ids) == 0 || formErrors["RedditSourceIDs"] != nil {
		return formErrors
	}
	db, err := getters.DBFromContext(r.Context())
	if err != nil {
		formErrors["RedditSourceIDs"] = err
		return formErrors
	}
	query := db.WithContext(r.Context()).Model(&RedditSource{}).Where("id IN ?", ids)
	if runner, ok := r.Context().Value("redditRunner").(RedditRunner); ok && runner.ID != 0 {
		query = query.Where("reddit_runner_id IS NULL OR reddit_runner_id = ?", runner.ID)
	} else {
		query = query.Where("reddit_runner_id IS NULL")
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		formErrors["RedditSourceIDs"] = err
		return formErrors
	}
	if count != int64(len(ids)) {
		formErrors["RedditSourceIDs"] = errors.New("select only Reddit sources without workers")
	}
	return formErrors
}
