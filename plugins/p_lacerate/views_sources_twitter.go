package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func init() {
	listPatchers := views.QueryPatchers[TwitterSource]{
		{Key: "lacerate.twitter_sources.preload_source", Value: views.QueryPatcherPreload[TwitterSource]{Field: "Source"}},
		{Key: "lacerate.twitter_sources.order_id", Value: views.QueryPatcherOrderBy[TwitterSource]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("lacerate.TwitterListView",
		lago.GetPageView("lacerate.TwitterSourcesTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.twitter_sources.list", views.LayerList[TwitterSource]{
				Key:           getters.Static("twitter_sources"),
				QueryPatchers: listPatchers,
			}))

	lago.RegistryView.Register("lacerate.TwitterDetailView",
		lago.GetPageView("lacerate.TwitterSourceDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.twitter_sources.detail", views.LayerDetail[TwitterSource]{
				Key:          getters.Static("twitterSource"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[TwitterSource]{
					{Key: "lacerate.twitter_sources.detail_preload", Value: views.QueryPatcherPreload[TwitterSource]{Field: "Source"}},
				},
			}))

	lago.RegistryView.Register("lacerate.TwitterCreateView",
		lago.GetPageView("lacerate.TwitterSourceCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.twitter_sources.create", twitterSourceCreateLayer{
				SuccessURL: lago.RoutePath("lacerate.TwitterDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("lacerate.TwitterUpdateView",
		lago.GetPageView("lacerate.TwitterSourceUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.twitter_sources.detail", views.LayerDetail[TwitterSource]{
				Key:          getters.Static("twitterSource"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[TwitterSource]{
					{Key: "lacerate.twitter_sources.update_preload", Value: views.QueryPatcherPreload[TwitterSource]{Field: "Source"}},
				},
			}).
			WithLayer("lacerate.twitter_sources.update", twitterSourceUpdateLayer{
				Key: getters.Static("twitterSource"),
				SuccessURL: lago.RoutePath("lacerate.TwitterDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("twitterSource.ID")),
				}),
			}))

	lago.RegistryView.Register("lacerate.TwitterDeleteView",
		lago.GetPageView("lacerate.TwitterSourceDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.twitter_sources.delete_detail", views.LayerDetail[TwitterSource]{
				Key:          getters.Static("twitterSource"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[TwitterSource]{
					{Key: "lacerate.twitter_sources.delete_preload", Value: views.QueryPatcherPreload[TwitterSource]{Field: "Source"}},
				},
			}).
			WithLayer("lacerate.twitter_sources.delete", twitterSourceDeleteLayer{
				Key:        getters.Static("twitterSource"),
				SuccessURL: lago.RoutePath("lacerate.TwitterDefaultRoute", nil),
			}))
}

type twitterSourceCreateLayer struct {
	SuccessURL getters.Getter[string]
}

func (m twitterSourceCreateLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("lacerate: twitter source create: parse form", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		name := strings.TrimSpace(fmt.Sprint(values["Name"]))
		if name == "" {
			fieldErrors["Name"] = fmt.Errorf("name is required")
		}

		var handlesJ datatypes.JSON
		if fieldErrors["Handles"] == nil {
			var ok bool
			handlesJ, ok = values["Handles"].(datatypes.JSON)
			if !ok {
				fieldErrors["Handles"] = fmt.Errorf("invalid handles value (got %T)", values["Handles"])
			} else if err := twitterAtLeastOneHandle(handlesJ); err != nil {
				fieldErrors["Handles"] = err
			}
		}

		if len(fieldErrors) != 0 {
			for fname, ferr := range fieldErrors {
				slog.Error("lacerate: twitter source create: field error", "field", fname, "error", ferr)
			}
			ctx = views.ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		duration := durationPtrFromFormValues(values)

		db := ctx.Value("$db").(*gorm.DB)
		var ts TwitterSource
		err = db.Transaction(func(tx *gorm.DB) error {
			src := Source{Name: name, Kind: "twitter", Duration: duration}
			if err := tx.Create(&src).Error; err != nil {
				return err
			}
			ts = TwitterSource{SourceID: src.ID, Handles: handlesJ}
			return tx.Create(&ts).Error
		})
		if err != nil {
			slog.Error("lacerate: twitter source create: transaction", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": fmt.Errorf("%v", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ctx = context.WithValue(ctx, "$id", ts.ID)
		if m.SuccessURL == nil {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successURL, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("lacerate: twitter source create: success URL", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

type twitterSourceUpdateLayer struct {
	Key        getters.Getter[string]
	SuccessURL getters.Getter[string]
}

func (m twitterSourceUpdateLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("lacerate: twitter source update: parse form", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("lacerate: twitter source update: context key", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_global": fmt.Errorf("failed to resolve context key: %w", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		record, ok := ctx.Value(key).(TwitterSource)
		if !ok {
			slog.Error("lacerate: twitter source update: missing record", "key", key)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_global": fmt.Errorf("record not found in context")})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		name := strings.TrimSpace(fmt.Sprint(values["Name"]))
		if name == "" {
			fieldErrors["Name"] = fmt.Errorf("name is required")
		}
		var handlesJ datatypes.JSON
		if fieldErrors["Handles"] == nil {
			var ok bool
			handlesJ, ok = values["Handles"].(datatypes.JSON)
			if !ok {
				fieldErrors["Handles"] = fmt.Errorf("invalid handles value (got %T)", values["Handles"])
			} else if err := twitterAtLeastOneHandle(handlesJ); err != nil {
				fieldErrors["Handles"] = err
			}
		}
		if len(fieldErrors) != 0 {
			for fname, ferr := range fieldErrors {
				slog.Error("lacerate: twitter source update: field error", "field", fname, "error", ferr)
			}
			ctx = views.ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		db := ctx.Value("$db").(*gorm.DB)
		tid := record.ID
		srcID := record.SourceID
		duration := durationPtrFromFormValues(values)

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&TwitterSource{}).Where("id = ?", tid).Updates(map[string]any{
				"Handles": handlesJ,
			}).Error; err != nil {
				return err
			}
			return tx.Model(&Source{Model: gorm.Model{ID: srcID}}).Updates(map[string]any{
				"Name":     name,
				"Duration": duration,
			}).Error
		})
		if err != nil {
			slog.Error("lacerate: twitter source update: transaction", "error", err)
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
			slog.Error("lacerate: twitter source update: success URL", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

func twitterAtLeastOneHandle(handlesJSON datatypes.JSON) error {
	if len(handlesJSON) == 0 {
		err := fmt.Errorf("at least one handle is required")
		slog.Error("lacerate: twitter handles validation", "error", err)
		return err
	}
	var arr []string
	if err := json.Unmarshal(handlesJSON, &arr); err != nil {
		err = fmt.Errorf("handles must be a JSON array of strings: %w", err)
		slog.Error("lacerate: twitter handles validation", "error", err)
		return err
	}
	for _, s := range arr {
		if strings.TrimSpace(s) != "" {
			return nil
		}
	}
	err := fmt.Errorf("at least one handle is required")
	slog.Error("lacerate: twitter handles validation", "error", err)
	return err
}

type twitterSourceDeleteLayer struct {
	Key        getters.Getter[string]
	SuccessURL getters.Getter[string]
}

func (m twitterSourceDeleteLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("lacerate: twitter source delete: context key", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("failed to resolve context key: %w", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		record, ok := ctx.Value(key).(TwitterSource)
		if !ok {
			slog.Error("lacerate: twitter source delete: missing record", "key", key)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("record not found in context")})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		srcID := record.SourceID
		tid := record.ID
		db := ctx.Value("$db").(*gorm.DB)

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Delete(&TwitterSource{}, tid).Error; err != nil {
				return err
			}
			return tx.Delete(&Source{Model: gorm.Model{ID: srcID}}).Error
		})
		if err != nil {
			slog.Error("lacerate: twitter source delete", "error", err)
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
			slog.Error("lacerate: twitter source delete: success URL", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}
