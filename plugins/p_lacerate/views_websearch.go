package p_lacerate

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type websearchRunAfterCreateLayer struct{}

func (websearchRunAfterCreateLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		id, err := getters.Key[uint]("$id")(ctx)
		if err != nil || id == 0 {
			slog.Error("lacerate: websearch create missing id", "error", err)
			next.ServeHTTP(w, r)
			return
		}
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("lacerate: websearch create db from context", "error", dberr)
			next.ServeHTTP(w, r)
			return
		}
		scheduleWebsearchRun(db, id)
		next.ServeHTTP(w, r)
	})
}

type websearchRunAfterUpdateLayer struct{}

func (websearchRunAfterUpdateLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		row, ok := ctx.Value("websearch").(Websearch)
		if !ok || row.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("lacerate: websearch update db from context", "error", dberr)
			next.ServeHTTP(w, r)
			return
		}
		scheduleWebsearchRun(db, row.ID)
		next.ServeHTTP(w, r)
	})
}

type websearchCreateRedirectLayer struct{}

func (websearchCreateRedirectLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		if errs, ok := ctx.Value(getters.ContextKeyError).(map[string]error); ok && len(errs) != 0 {
			next.ServeHTTP(w, r)
			return
		}
		id, err := getters.Key[uint]("$id")(ctx)
		if err != nil || id == 0 {
			next.ServeHTTP(w, r)
			return
		}
		successURL, err := lago.RoutePath("lacerate.WebsearchDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Key[uint]("$id")),
		})(ctx)
		if err != nil {
			slog.Error("lacerate: websearch create redirect URL", "error", err, "websearch_id", id)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

type websearchUpdateRedirectLayer struct{}

func (websearchUpdateRedirectLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		if errs, ok := ctx.Value(getters.ContextKeyError).(map[string]error); ok && len(errs) != 0 {
			next.ServeHTTP(w, r)
			return
		}
		row, ok := ctx.Value("websearch").(Websearch)
		if !ok || row.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}
		successURL, err := lago.RoutePath("lacerate.WebsearchDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Key[uint]("websearch.ID")),
		})(ctx)
		if err != nil {
			slog.Error("lacerate: websearch update redirect URL", "error", err, "websearch_id", row.ID)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

type websearchRelatedIntelLayer struct{}

func (websearchRelatedIntelLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		row, ok := ctx.Value("websearch").(Websearch)
		if !ok || row.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("lacerate: websearch related intel db from context", "error", dberr)
			next.ServeHTTP(w, r)
			return
		}
		var links []WebsearchIntel
		if err := db.WithContext(ctx).
			Preload("Intel.Source").
			Where("websearch_id = ?", row.ID).
			Order("id DESC").
			Find(&links).Error; err != nil {
			slog.Error("lacerate: websearch related intel load", "error", err, "websearch_id", row.ID)
			next.ServeHTTP(w, r)
			return
		}
		items := make([]Intel, 0, len(links))
		for i := range links {
			if links[i].Intel.ID != 0 {
				items = append(items, links[i].Intel)
			}
		}
		ctx = context.WithValue(ctx, ctxKeyWebsearchRelatedIntel, components.ObjectList[Intel]{
			Items:    items,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(items)),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type websearchDeleteIntelLayer struct{}

func (websearchDeleteIntelLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		row, ok := ctx.Value("websearch").(Websearch)
		if !ok || row.ID == 0 {
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("websearch record missing")})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("lacerate: websearch delete intel db from context", "error", dberr)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": dberr})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		intelIDStr := strings.TrimSpace(r.PathValue("intel_id"))
		intelID64, err := strconv.ParseUint(intelIDStr, 10, 32)
		if err != nil || intelID64 == 0 {
			err = fmt.Errorf("invalid intel id %q", intelIDStr)
			slog.Error("lacerate: websearch delete intel parse path", "error", err, "websearch_id", row.ID, "intel_id", intelIDStr)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		intelID := uint(intelID64)

		err = db.Transaction(func(tx *gorm.DB) error {
			var link WebsearchIntel
			if err := tx.Where("websearch_id = ? AND intel_id = ?", row.ID, intelID).First(&link).Error; err != nil {
				return err
			}
			return tx.Where("id = ?", intelID).Delete(&Intel{}).Error
		})
		if err != nil {
			slog.Error("lacerate: websearch delete intel", "error", err, "websearch_id", row.ID, "intel_id", intelID)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		successURL, err := lago.RoutePath("lacerate.WebsearchDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Key[uint]("websearch.ID")),
		})(ctx)
		if err != nil {
			slog.Error("lacerate: websearch delete intel success URL", "error", err, "websearch_id", row.ID)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

type websearchDeleteLayer struct{}

func (websearchDeleteLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		row, ok := ctx.Value("websearch").(Websearch)
		if !ok || row.ID == 0 {
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("websearch record missing")})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("lacerate: websearch delete db from context", "error", dberr)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": dberr})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("websearch_id = ?", row.ID).Delete(&WebsearchIntel{}).Error; err != nil {
				return err
			}
			return tx.Delete(&Websearch{Model: gorm.Model{ID: row.ID}}).Error
		})
		if err != nil {
			slog.Error("lacerate: websearch delete", "error", err, "websearch_id", row.ID)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("failed to delete record: %w", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		successURL, err := lago.RoutePath("lacerate.WebsearchListRoute", nil)(ctx)
		if err != nil {
			slog.Error("lacerate: websearch delete success URL", "error", err, "websearch_id", row.ID)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

func runAndStoreWebsearchIntel(ctx context.Context, db *gorm.DB, row Websearch) error {
	query := strings.TrimSpace(row.Query)
	if query == "" {
		return fmt.Errorf("query is required")
	}

	var existingHashRows []string
	if err := db.WithContext(ctx).
		Model(&Intel{}).
		Joins("JOIN websearch_intels ON websearch_intels.intel_id = intels.id").
		Where("websearch_intels.websearch_id = ? AND intels.dedup_hash IS NOT NULL AND intels.dedup_hash <> ''", row.ID).
		Pluck("intels.dedup_hash", &existingHashRows).Error; err != nil {
		return err
	}
	existingDedup := make(map[string]struct{}, len(existingHashRows))
	for _, h := range existingHashRows {
		existingDedup[h] = struct{}{}
	}

	intels, err := runWebsearchQueryFetch(ctx, db.WithContext(ctx), nil, query, existingDedup)
	if err != nil {
		return err
	}
	if len(intels) == 0 {
		return nil
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range intels {
			if strings.TrimSpace(intels[i].Content) == "" {
				continue
			}
			intels[i].SourceID = nil
			if err := tx.Create(&intels[i]).Error; err != nil {
				return err
			}
			if err := tx.Create(&WebsearchIntel{
				WebsearchID: row.ID,
				IntelID:     intels[i].ID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

var (
	websearchRunMu      sync.Mutex
	websearchRunsActive = map[uint]struct{}{}
)

func scheduleWebsearchRun(db *gorm.DB, websearchID uint) {
	if db == nil || websearchID == 0 {
		return
	}
	websearchRunMu.Lock()
	if _, ok := websearchRunsActive[websearchID]; ok {
		websearchRunMu.Unlock()
		return
	}
	websearchRunsActive[websearchID] = struct{}{}
	websearchRunMu.Unlock()

	go func() {
		defer func() {
			websearchRunMu.Lock()
			delete(websearchRunsActive, websearchID)
			websearchRunMu.Unlock()
		}()

		startedAt := time.Now().UTC()
		if err := db.Model(&Websearch{}).Where("id = ?", websearchID).Updates(map[string]any{
			"status":          "running",
			"last_run_error":  "",
			"last_started_at": startedAt,
		}).Error; err != nil {
			slog.Error("lacerate: websearch set running status", "error", err, "websearch_id", websearchID)
			return
		}

		var row Websearch
		if err := db.First(&row, websearchID).Error; err != nil {
			slog.Error("lacerate: websearch run load row", "error", err, "websearch_id", websearchID)
			return
		}

		runErr := runAndStoreWebsearchIntel(context.Background(), db, row)
		endedAt := time.Now().UTC()
		updates := map[string]any{
			"last_ended_at": endedAt,
		}
		if runErr != nil {
			updates["status"] = "failed"
			updates["last_run_error"] = runErr.Error()
			slog.Error("lacerate: websearch background run", "error", runErr, "websearch_id", websearchID)
		} else {
			updates["status"] = "done"
			updates["last_run_error"] = ""
		}
		if err := db.Model(&Websearch{}).Where("id = ?", websearchID).Updates(updates).Error; err != nil {
			slog.Error("lacerate: websearch set final status", "error", err, "websearch_id", websearchID)
		}
	}()
}

func init() {
	patchers := views.QueryPatchers[Websearch]{
		{Key: "lacerate.websearch.order_id", Value: views.QueryPatcherOrderBy[Websearch]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("lacerate.WebsearchListView",
		lago.GetPageView("lacerate.WebsearchTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.websearch.list", views.LayerList[Websearch]{
				Key:           getters.Static("websearches"),
				QueryPatchers: patchers,
			}))

	lago.RegistryView.Register("lacerate.WebsearchCreateView",
		lago.GetPageView("lacerate.WebsearchCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.websearch.create", views.LayerCreate[Websearch]{
				SuccessURL: nil,
			}).
			WithLayer("lacerate.websearch.create_run", websearchRunAfterCreateLayer{}).
			WithLayer("lacerate.websearch.create_redirect", websearchCreateRedirectLayer{}))

	lago.RegistryView.Register("lacerate.WebsearchDetailView",
		lago.GetPageView("lacerate.WebsearchDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.websearch.detail", views.LayerDetail[Websearch]{
				Key:          getters.Static("websearch"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.websearch.related_intel", websearchRelatedIntelLayer{}))

	lago.RegistryView.Register("lacerate.WebsearchUpdateView",
		lago.GetPageView("lacerate.WebsearchUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.websearch.update_detail", views.LayerDetail[Websearch]{
				Key:          getters.Static("websearch"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.websearch.update", views.LayerUpdate[Websearch]{
				Key:        getters.Static("websearch"),
				SuccessURL: nil,
			}).
			WithLayer("lacerate.websearch.update_run", websearchRunAfterUpdateLayer{}).
			WithLayer("lacerate.websearch.update_redirect", websearchUpdateRedirectLayer{}))

	lago.RegistryView.Register("lacerate.WebsearchDeleteView",
		lago.GetPageView("lacerate.WebsearchDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.websearch.delete_detail", views.LayerDetail[Websearch]{
				Key:          getters.Static("websearch"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.websearch.delete", websearchDeleteLayer{}))

	lago.RegistryView.Register("lacerate.WebsearchDeleteIntelView",
		lago.GetPageView("lacerate.WebsearchDeleteIntelForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.websearch.delete_intel_path", views.PathLayer{Names: []string{"intel_id"}}).
			WithLayer("lacerate.websearch.delete_intel_detail", views.LayerDetail[Websearch]{
				Key:          getters.Static("websearch"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.websearch.delete_intel", websearchDeleteIntelLayer{}))
}
