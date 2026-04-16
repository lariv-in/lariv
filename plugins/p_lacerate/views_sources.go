package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// sourceWorkerPOSTOnlyLayer rejects non-POST for worker action routes.
type sourceWorkerPOSTOnlyLayer struct{}

func (sourceWorkerPOSTOnlyLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func sourceRestartWorkerHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		id64, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("lacerate: source restart worker missing db", "source_id", idStr, "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		ScheduleRestartSourceWorker(db, uint(id64))
		detailURL, err := lago.RoutePath("lacerate.SourceDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("lacerate: source restart worker detail URL", "error", err, "source_id", idStr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

func sourceStopWorkerHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		id64, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		StopSourceWorker(uint(id64))
		detailURL, err := lago.RoutePath("lacerate.SourceDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("lacerate: source stop worker detail URL", "error", err, "source_id", idStr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

func init() {
	lago.RegistryView.Register("lacerate.SourceListView",
		lago.GetPageView("lacerate.SourcesTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.sources.list", sourceListLayer{}))

	lago.RegistryView.Register("lacerate.SourceDetailView",
		lago.GetPageView("lacerate.SourceDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.sources.detail", sourceDetailLayer{
				Key:          getters.Static("sourcePageData"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("lacerate.SourceCreateView",
		lago.GetPageView("lacerate.SourceCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.sources.create", sourceCreateLayer{
				SuccessURL: lago.RoutePath("lacerate.SourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("lacerate.SourceUpdateView",
		lago.GetPageView("lacerate.SourceUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.sources.update_detail", sourceDetailLayer{
				Key:          getters.Static("sourcePageData"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.sources.update", sourceUpdateLayer{
				Key: getters.Static("sourcePageData"),
				SuccessURL: lago.RoutePath("lacerate.SourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("sourcePageData.Source.ID")),
				}),
			}))

	lago.RegistryView.Register("lacerate.SourceDeleteView",
		lago.GetPageView("lacerate.SourceDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.sources.delete_detail", sourceDetailLayer{
				Key:          getters.Static("sourcePageData"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.sources.delete", sourceDeleteLayer{
				Key:        getters.Static("sourcePageData"),
				SuccessURL: lago.RoutePath("lacerate.SourceListRoute", nil),
			}))

	lago.RegistryView.Register("lacerate.SourceRestartWorkerView",
		lago.GetPageView("lacerate.SourceDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.sources.restart_worker_post_only", sourceWorkerPOSTOnlyLayer{}).
			WithLayer("lacerate.sources.restart_worker", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: sourceRestartWorkerHandler,
			}))

	lago.RegistryView.Register("lacerate.SourceStopWorkerView",
		lago.GetPageView("lacerate.SourceDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.sources.stop_worker_post_only", sourceWorkerPOSTOnlyLayer{}).
			WithLayer("lacerate.sources.stop_worker", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: sourceStopWorkerHandler,
			}))
}

type sourceListLayer struct{}

func (sourceListLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("lacerate: db from context", "error", dberr)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": dberr})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		var sources []Source
		if err := db.WithContext(ctx).Order("id DESC").Find(&sources).Error; err != nil {
			slog.Error("lacerate: source list load", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		items, err := loadSourcePageDataList(ctx, db, sources)
		if err != nil {
			slog.Error("lacerate: source list page data", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, "sources", components.ObjectList[SourcePageData]{
			Items:    items,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(items)),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type sourceDetailLayer struct {
	Key, PathParamKey getters.Getter[string]
}

func (m sourceDetailLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		pathParamKey, err := m.PathParamKey(ctx)
		if err != nil {
			slog.Error("lacerate: source detail path key", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		idStr := r.PathValue(pathParamKey)
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			slog.Error("lacerate: source detail parse id", "error", err, "id", idStr)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("invalid ID %q", idStr)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("lacerate: db from context", "error", dberr)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": dberr})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		data, err := loadSourcePageData(ctx, db, uint(id))
		if err != nil {
			slog.Error("lacerate: source detail load", "error", err, "id", id)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("record not found")})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("lacerate: source detail context key", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, key, data)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type sourceCreateLayer struct {
	SuccessURL getters.Getter[string]
}

func (m sourceCreateLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("lacerate: source create parse form", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		formData := sourceFormDataFromValues(values, fieldErrors)
		if len(fieldErrors) != 0 {
			logSourceFieldErrors("create", fieldErrors)
			ctx = views.ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("lacerate: db from context", "error", dberr)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": dberr})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		var sourceID uint
		err = db.Transaction(func(tx *gorm.DB) error {
			src := Source{Name: formData.Name, Kind: formData.Kind, Duration: formData.Duration}
			if err := tx.Create(&src).Error; err != nil {
				return err
			}
			sourceID = src.ID
			return createSourceKindRow(tx, sourceID, formData)
		})
		if err != nil {
			slog.Error("lacerate: source create transaction", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": fmt.Errorf("%v", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ScheduleRestartSourceWorker(db, sourceID)

		ctx = context.WithValue(ctx, "$id", sourceID)
		successURL, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("lacerate: source create success URL", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

type sourceUpdateLayer struct {
	Key        getters.Getter[string]
	SuccessURL getters.Getter[string]
}

func (m sourceUpdateLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("lacerate: source update parse form", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("lacerate: source update context key", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		record, ok := ctx.Value(key).(SourcePageData)
		if !ok {
			slog.Error("lacerate: source update missing record", "key", key)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_global": fmt.Errorf("record not found in context")})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		formData := sourceFormDataFromValues(values, fieldErrors)
		if len(fieldErrors) != 0 {
			logSourceFieldErrors("update", fieldErrors)
			ctx = views.ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("lacerate: db from context", "error", dberr)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": dberr})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		sourceID := record.Source.ID
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&Source{Model: gorm.Model{ID: sourceID}}).Updates(map[string]any{
				"Name":     formData.Name,
				"Kind":     formData.Kind,
				"Duration": formData.Duration,
			}).Error; err != nil {
				return err
			}
			if err := deleteSourceKindExtensionRows(tx, sourceID); err != nil {
				return err
			}
			return createSourceKindRow(tx, sourceID, formData)
		})
		if err != nil {
			slog.Error("lacerate: source update transaction", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": fmt.Errorf("%v", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ScheduleRestartSourceWorker(db, sourceID)

		successURL, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("lacerate: source update success URL", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

type sourceDeleteLayer struct {
	Key        getters.Getter[string]
	SuccessURL getters.Getter[string]
}

func (m sourceDeleteLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("lacerate: source delete context key", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		record, ok := ctx.Value(key).(SourcePageData)
		if !ok {
			slog.Error("lacerate: source delete missing record", "key", key)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("record not found in context")})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		sourceID := record.Source.ID
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("lacerate: db from context", "error", dberr)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": dberr})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := deleteSourceKindExtensionRows(tx, sourceID); err != nil {
				return err
			}
			if err := tx.Model(&Intel{}).Where("source_id = ?", sourceID).Update("source_id", nil).Error; err != nil {
				return err
			}
			return tx.Delete(&Source{Model: gorm.Model{ID: sourceID}}).Error
		})
		if err != nil {
			slog.Error("lacerate: source delete", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("failed to delete: %w", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successURL, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("lacerate: source delete success URL", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

type parsedSourceFormData struct {
	Name          string
	Kind          string
	Duration      time.Duration
	Subreddits    datatypes.JSON
	SearchQuery   string
	MaxFreshPosts uint
	Handles       datatypes.JSON
	URL           string
	Query         string
}

func parseSourceMaxFreshPostsInto(values map[string]any, fieldErrors map[string]error, out *uint) {
	if v, ok := values["MaxFreshPosts"].(uint); ok {
		*out = v
		return
	}
	if values["MaxFreshPosts"] != nil {
		fieldErrors["MaxFreshPosts"] = fmt.Errorf("invalid max fresh posts value (got %T)", values["MaxFreshPosts"])
	}
}

func sourceFormDataFromValues(values map[string]any, fieldErrors map[string]error) parsedSourceFormData {
	out := parsedSourceFormData{
		Name:     strings.TrimSpace(fmt.Sprint(values["Name"])),
		Kind:     strings.TrimSpace(fmt.Sprint(values["Kind"])),
		Duration: durationPtrFromFormValues(values),
	}
	if out.Name == "" {
		fieldErrors["Name"] = fmt.Errorf("name is required")
	}
	if _, ok := registry.PairFromPairs(out.Kind, SourceKindChoices); !ok {
		fieldErrors["Kind"] = fmt.Errorf("kind is required")
		return out
	}
	switch out.Kind {
	case "reddit":
		out.SearchQuery = strings.TrimSpace(fmt.Sprint(values["SearchQuery"]))
		parseSourceMaxFreshPostsInto(values, fieldErrors, &out.MaxFreshPosts)
		if fieldErrors["Subreddits"] == nil {
			subreddits, ok := values["Subreddits"].(datatypes.JSON)
			if !ok {
				fieldErrors["Subreddits"] = fmt.Errorf("invalid subreddits value (got %T)", values["Subreddits"])
			} else {
				out.Subreddits = subreddits
			}
		}
	case "twitter":
		parseSourceMaxFreshPostsInto(values, fieldErrors, &out.MaxFreshPosts)
		if fieldErrors["Handles"] == nil {
			handles, ok := values["Handles"].(datatypes.JSON)
			if !ok {
				fieldErrors["Handles"] = fmt.Errorf("invalid handles value (got %T)", values["Handles"])
			} else if err := twitterAtLeastOneHandle(handles); err != nil {
				fieldErrors["Handles"] = err
			} else {
				out.Handles = handles
			}
		}
	case "website":
		rawURL := strings.TrimSpace(fmt.Sprint(values["URL"]))
		if rawURL == "" {
			fieldErrors["URL"] = fmt.Errorf("url is required")
			return out
		}
		normalizedURL, err := normalizeWebsiteSeedURL(rawURL)
		if err != nil {
			fieldErrors["URL"] = err
			return out
		}
		out.URL = normalizedURL
	case sourceKindGoogleSearch:
		out.Query = strings.TrimSpace(fmt.Sprint(values["Query"]))
		if out.Query == "" {
			fieldErrors["Query"] = fmt.Errorf("query is required")
		}
	case sourceKindWebsearch:
		out.Query = strings.TrimSpace(fmt.Sprint(values["Query"]))
		if out.Query == "" {
			fieldErrors["Query"] = fmt.Errorf("query is required")
		}
	case sourceKindDirectMedia:
		rawURL := strings.TrimSpace(fmt.Sprint(values["URL"]))
		if rawURL == "" {
			fieldErrors["URL"] = fmt.Errorf("url is required")
			return out
		}
		normalizedURL, err := normalizeWebsiteSeedURL(rawURL)
		if err != nil {
			fieldErrors["URL"] = err
			return out
		}
		out.URL = normalizedURL
	}
	return out
}

func createSourceKindRow(tx *gorm.DB, sourceID uint, formData parsedSourceFormData) error {
	switch formData.Kind {
	case "reddit":
		return tx.Create(&RedditSource{
			SourceID:      sourceID,
			Subreddits:    formData.Subreddits,
			SearchQuery:   formData.SearchQuery,
			MaxFreshPosts: sourceMaxFreshPostsForSave(formData.MaxFreshPosts),
		}).Error
	case "twitter":
		return tx.Create(&TwitterSource{
			SourceID:      sourceID,
			Handles:       formData.Handles,
			MaxFreshPosts: sourceMaxFreshPostsForSave(formData.MaxFreshPosts),
		}).Error
	case "website":
		return tx.Create(&WebsiteSource{
			SourceID: sourceID,
			URL:      formData.URL,
		}).Error
	case sourceKindGoogleSearch:
		return tx.Create(&GoogleSearchSource{
			SourceID: sourceID,
			Query:    formData.Query,
		}).Error
	case sourceKindWebsearch:
		return tx.Create(&WebsearchSource{
			SourceID: sourceID,
			Query:    formData.Query,
		}).Error
	case sourceKindDirectMedia:
		return tx.Create(&DirectMediaSource{
			SourceID: sourceID,
			URL:      formData.URL,
		}).Error
	default:
		return fmt.Errorf("unsupported source kind %q", formData.Kind)
	}
}

func logSourceFieldErrors(action string, fieldErrors map[string]error) {
	for field, err := range fieldErrors {
		slog.Error("lacerate: source field error", "action", action, "field", field, "error", err)
	}
}

// durationPtrFromFormValues reads *time.Duration produced by [components.InputDuration.Parse].
func durationPtrFromFormValues(values map[string]any) time.Duration {
	dp, _ := values["Duration"].(*time.Duration)
	if dp == nil {
		return 0
	}
	return *dp
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
