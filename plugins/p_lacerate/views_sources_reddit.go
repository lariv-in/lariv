package p_lacerate

import (
	"context"
	"fmt"
	"log/slog"
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
	listPatchers := views.QueryPatchers[RedditSource]{
		{Key: "lacerate.reddit_sources.preload_source", Value: views.QueryPatcherPreload[RedditSource]{Field: "Source"}},
		{Key: "lacerate.reddit_sources.order_id", Value: views.QueryPatcherOrderBy[RedditSource]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("lacerate.ListView",
		lago.GetPageView("lacerate.RedditSourcesTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reddit_sources.list", views.LayerList[RedditSource]{
				Key:           getters.Static("reddit_sources"),
				QueryPatchers: listPatchers,
			}))

	lago.RegistryView.Register("lacerate.DetailView",
		lago.GetPageView("lacerate.RedditSourceDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reddit_sources.detail", views.LayerDetail[RedditSource]{
				Key:          getters.Static("redditSource"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[RedditSource]{
					{Key: "lacerate.reddit_sources.detail_preload", Value: views.QueryPatcherPreload[RedditSource]{Field: "Source"}},
				},
			}))

	lago.RegistryView.Register("lacerate.CreateView",
		lago.GetPageView("lacerate.RedditSourceCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reddit_sources.create", redditSourceCreateLayer{
				SuccessURL: lago.RoutePath("lacerate.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("lacerate.UpdateView",
		lago.GetPageView("lacerate.RedditSourceUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reddit_sources.detail", views.LayerDetail[RedditSource]{
				Key:          getters.Static("redditSource"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[RedditSource]{
					{Key: "lacerate.reddit_sources.update_preload", Value: views.QueryPatcherPreload[RedditSource]{Field: "Source"}},
				},
			}).
			WithLayer("lacerate.reddit_sources.update", redditSourceUpdateLayer{
				Key: getters.Static("redditSource"),
				SuccessURL: lago.RoutePath("lacerate.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditSource.ID")),
				}),
			}))

	lago.RegistryView.Register("lacerate.DeleteView",
		lago.GetPageView("lacerate.RedditSourceDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reddit_sources.delete_detail", views.LayerDetail[RedditSource]{
				Key:          getters.Static("redditSource"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[RedditSource]{
					{Key: "lacerate.reddit_sources.delete_preload", Value: views.QueryPatcherPreload[RedditSource]{Field: "Source"}},
				},
			}).
			WithLayer("lacerate.reddit_sources.delete", redditSourceDeleteLayer{
				Key:        getters.Static("redditSource"),
				SuccessURL: lago.RoutePath("lacerate.DefaultRoute", nil),
			}))
}

// redditSourceCreateLayer creates a Source row (Kind=reddit) and a RedditSource in one transaction.
type redditSourceCreateLayer struct {
	SuccessURL getters.Getter[string]
}

func (m redditSourceCreateLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("lacerate: reddit source create: parse form", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		name := strings.TrimSpace(fmt.Sprint(values["Name"]))
		if name == "" {
			fieldErrors["Name"] = fmt.Errorf("name is required")
		}

		var subJ datatypes.JSON
		if fieldErrors["Subreddits"] == nil {
			var ok bool
			subJ, ok = values["Subreddits"].(datatypes.JSON)
			if !ok {
				fieldErrors["Subreddits"] = fmt.Errorf("invalid subreddits value (got %T)", values["Subreddits"])
			}
		}

		if len(fieldErrors) != 0 {
			for fname, ferr := range fieldErrors {
				slog.Error("lacerate: reddit source create: field error", "field", fname, "error", ferr)
			}
			ctx = views.ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		searchQuery := strings.TrimSpace(fmt.Sprint(values["SearchQuery"]))
		duration := durationPtrFromFormValues(values)

		db := ctx.Value("$db").(*gorm.DB)
		var rs RedditSource
		err = db.Transaction(func(tx *gorm.DB) error {
			src := Source{Name: name, Kind: "reddit", Duration: duration}
			if err := tx.Create(&src).Error; err != nil {
				return err
			}
			rs = RedditSource{SourceID: src.ID, Subreddits: subJ, SearchQuery: searchQuery}
			return tx.Create(&rs).Error
		})
		if err != nil {
			slog.Error("lacerate: reddit source create: transaction", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": fmt.Errorf("%v", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ctx = context.WithValue(ctx, "$id", rs.ID)
		if m.SuccessURL == nil {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successURL, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("lacerate: reddit source create: success URL", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

// redditSourceUpdateLayer updates RedditSource.Subreddits and the parent Source.Name.
type redditSourceUpdateLayer struct {
	Key        getters.Getter[string]
	SuccessURL getters.Getter[string]
}

func (m redditSourceUpdateLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("lacerate: reddit source update: parse form", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("lacerate: reddit source update: context key", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_global": fmt.Errorf("failed to resolve context key: %w", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		record, ok := ctx.Value(key).(RedditSource)
		if !ok {
			slog.Error("lacerate: reddit source update: missing record", "key", key)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_global": fmt.Errorf("record not found in context")})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		name := strings.TrimSpace(fmt.Sprint(values["Name"]))
		if name == "" {
			fieldErrors["Name"] = fmt.Errorf("name is required")
		}
		var subJ datatypes.JSON
		if fieldErrors["Subreddits"] == nil {
			var ok bool
			subJ, ok = values["Subreddits"].(datatypes.JSON)
			if !ok {
				fieldErrors["Subreddits"] = fmt.Errorf("invalid subreddits value (got %T)", values["Subreddits"])
			}
		}
		if len(fieldErrors) != 0 {
			for fname, ferr := range fieldErrors {
				slog.Error("lacerate: reddit source update: field error", "field", fname, "error", ferr)
			}
			ctx = views.ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		db := ctx.Value("$db").(*gorm.DB)
		rid := record.ID
		srcID := record.SourceID
		searchQuery := strings.TrimSpace(fmt.Sprint(values["SearchQuery"]))
		duration := durationPtrFromFormValues(values)

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&RedditSource{}).Where("id = ?", rid).Updates(map[string]any{
				"Subreddits":  subJ,
				"SearchQuery": searchQuery,
			}).Error; err != nil {
				return err
			}
			return tx.Model(&Source{Model: gorm.Model{ID: srcID}}).Updates(map[string]any{
				"Name":     name,
				"Duration": duration,
			}).Error
		})
		if err != nil {
			slog.Error("lacerate: reddit source update: transaction", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": fmt.Errorf("%v", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if m.SuccessURL == nil {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successURL, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("lacerate: reddit source update: success URL", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

type redditSourceDeleteLayer struct {
	Key        getters.Getter[string]
	SuccessURL getters.Getter[string]
}

func (m redditSourceDeleteLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("lacerate: reddit source delete: context key", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("failed to resolve context key: %w", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		record, ok := ctx.Value(key).(RedditSource)
		if !ok {
			slog.Error("lacerate: reddit source delete: missing record", "key", key)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("record not found in context")})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		srcID := record.SourceID
		rid := record.ID
		db := ctx.Value("$db").(*gorm.DB)

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Delete(&RedditSource{}, rid).Error; err != nil {
				return err
			}
			return tx.Delete(&Source{Model: gorm.Model{ID: srcID}}).Error
		})
		if err != nil {
			slog.Error("lacerate: reddit source delete", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("failed to delete: %w", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if m.SuccessURL == nil {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successURL, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("lacerate: reddit source delete: success URL", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

// durationPtrFromFormValues reads *time.Duration produced by [components.InputDuration.Parse].
func durationPtrFromFormValues(values map[string]any) time.Duration {
	dp, _ := values["Duration"].(*time.Duration)
	if dp == nil {
		return 0
	}
	return *dp
}
