package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryView.Register("lacerate.ReportListView",
		lago.GetPageView("lacerate.ReportsTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reports.list", reportListLayer{}))

	lago.RegistryView.Register("lacerate.ReportDetailView",
		lago.GetPageView("lacerate.ReportDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reports.detail", reportDetailLayer{
				Key:          getters.Static("reportPageData"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("lacerate.ReportCreateView",
		lago.GetPageView("lacerate.ReportCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reports.create", reportCreateLayer{
				SuccessURL: lago.RoutePath("lacerate.ReportDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("lacerate.ReportUpdateView",
		lago.GetPageView("lacerate.ReportUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reports.update_detail", reportDetailLayer{
				Key:          getters.Static("reportPageData"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.reports.update", reportUpdateLayer{
				Key: getters.Static("reportPageData"),
				SuccessURL: lago.RoutePath("lacerate.ReportDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("reportPageData.Report.ID")),
				}),
			}))

	lago.RegistryView.Register("lacerate.ReportDeleteView",
		lago.GetPageView("lacerate.ReportDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.reports.delete_detail", reportDetailLayer{
				Key:          getters.Static("reportPageData"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.reports.delete", reportDeleteLayer{
				Key:        getters.Static("reportPageData"),
				SuccessURL: lago.RoutePath("lacerate.ReportListRoute", nil),
			}))
}

type reportListLayer struct{}

func (reportListLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		db := ctx.Value("$db").(*gorm.DB)
		var reports []Report
		if err := db.WithContext(ctx).Order("id DESC").Find(&reports).Error; err != nil {
			slog.Error("lacerate: report list load", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		items, err := loadReportPageDataList(ctx, db, reports)
		if err != nil {
			slog.Error("lacerate: report list page data", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, "reports", components.ObjectList[ReportPageData]{
			Items:    items,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(items)),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type reportDetailLayer struct {
	Key, PathParamKey getters.Getter[string]
}

func (m reportDetailLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		pathParamKey, err := m.PathParamKey(ctx)
		if err != nil {
			slog.Error("lacerate: report detail path key", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		idStr := r.PathValue(pathParamKey)
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			slog.Error("lacerate: report detail parse id", "error", err, "id", idStr)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("invalid ID %q", idStr)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db := ctx.Value("$db").(*gorm.DB)
		data, err := loadReportPageData(ctx, db, uint(id))
		if err != nil {
			slog.Error("lacerate: report detail load", "error", err, "id", id)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("record not found")})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("lacerate: report detail context key", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, key, data)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type reportCreateLayer struct {
	SuccessURL getters.Getter[string]
}

func (m reportCreateLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("lacerate: report create parse form", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		formData := reportFormDataFromValues(ctx, values, fieldErrors)
		if len(fieldErrors) != 0 {
			logReportFieldErrors("create", fieldErrors)
			ctx = views.ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		db := ctx.Value("$db").(*gorm.DB)
		var reportID uint
		err = db.Transaction(func(tx *gorm.DB) error {
			report := Report{Name: formData.Name, Description: formData.Description, Kind: formData.Kind}
			if err := tx.Create(&report).Error; err != nil {
				return err
			}
			reportID = report.ID
			if err := createReportKindRow(tx, reportID, formData); err != nil {
				return err
			}
			return refreshReportEmbedding(ctx, tx, reportID)
		})
		if err != nil {
			slog.Error("lacerate: report create transaction", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": fmt.Errorf("%v", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, "$id", reportID)
		successURL, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("lacerate: report create success URL", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

type reportUpdateLayer struct {
	Key        getters.Getter[string]
	SuccessURL getters.Getter[string]
}

func (m reportUpdateLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("lacerate: report update parse form", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("lacerate: report update context key", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		record, ok := ctx.Value(key).(ReportPageData)
		if !ok {
			slog.Error("lacerate: report update missing record", "key", key)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_global": fmt.Errorf("record not found in context")})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		formData := reportFormDataFromValues(ctx, values, fieldErrors)
		if len(fieldErrors) != 0 {
			logReportFieldErrors("update", fieldErrors)
			ctx = views.ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db := ctx.Value("$db").(*gorm.DB)
		reportID := record.Report.ID
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&Report{Model: gorm.Model{ID: reportID}}).Updates(map[string]any{
				"Name":        formData.Name,
				"Description": formData.Description,
				"Kind":        formData.Kind,
			}).Error; err != nil {
				return err
			}
			if err := deleteReportKindExtensionRows(tx, reportID); err != nil {
				return err
			}
			if err := createReportKindRow(tx, reportID, formData); err != nil {
				return err
			}
			return refreshReportEmbedding(ctx, tx, reportID)
		})
		if err != nil {
			slog.Error("lacerate: report update transaction", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": fmt.Errorf("%v", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successURL, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("lacerate: report update success URL", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, values, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

type reportDeleteLayer struct {
	Key        getters.Getter[string]
	SuccessURL getters.Getter[string]
}

func (m reportDeleteLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("lacerate: report delete context key", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		record, ok := ctx.Value(key).(ReportPageData)
		if !ok {
			slog.Error("lacerate: report delete missing record", "key", key)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("record not found in context")})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		reportID := record.Report.ID
		db := ctx.Value("$db").(*gorm.DB)
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := deleteReportKindExtensionRows(tx, reportID); err != nil {
				return err
			}
			return tx.Delete(&Report{Model: gorm.Model{ID: reportID}}).Error
		})
		if err != nil {
			slog.Error("lacerate: report delete", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("failed to delete: %w", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successURL, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("lacerate: report delete success URL", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, successURL, http.StatusSeeOther)
	})
}

type reportTimelineEntryFormInput struct {
	Datetime string `json:"datetime"`
	Title    string `json:"title"`
	Content  string `json:"content"`
}

type parsedReportTimelineEntry struct {
	Datetime time.Time
	Title    string
	Content  string
}

type parsedReportFormData struct {
	Name                string
	Description         string
	Kind                string
	BriefingContent     string
	TimelineEntriesJSON string
	TimelineEntries     []parsedReportTimelineEntry
}

func parseTimelineEntriesJSON(ctx context.Context, raw string) ([]parsedReportTimelineEntry, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("timeline entries are required")
	}
	var input []reportTimelineEntryFormInput
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		return nil, fmt.Errorf("timeline entries must be a JSON array of objects: %w", err)
	}
	if len(input) == 0 {
		return nil, fmt.Errorf("at least one timeline entry is required")
	}
	tz, _ := ctx.Value("$tz").(*time.Location)
	if tz == nil {
		tz = time.UTC
	}
	out := make([]parsedReportTimelineEntry, 0, len(input))
	for i, row := range input {
		title := strings.TrimSpace(row.Title)
		content := strings.TrimSpace(row.Content)
		if title == "" {
			return nil, fmt.Errorf("timeline entry %d title is required", i+1)
		}
		if content == "" {
			return nil, fmt.Errorf("timeline entry %d content is required", i+1)
		}
		dtRaw := strings.TrimSpace(row.Datetime)
		if dtRaw == "" {
			return nil, fmt.Errorf("timeline entry %d datetime is required", i+1)
		}
		dt, err := time.ParseInLocation("2006-01-02T15:04", dtRaw, tz)
		if err != nil {
			return nil, fmt.Errorf("timeline entry %d datetime must use YYYY-MM-DDTHH:MM", i+1)
		}
		out = append(out, parsedReportTimelineEntry{
			Datetime: dt,
			Title:    title,
			Content:  content,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Datetime.Before(out[j].Datetime)
	})
	return out, nil
}

func reportFormDataFromValues(ctx context.Context, values map[string]any, fieldErrors map[string]error) parsedReportFormData {
	out := parsedReportFormData{
		Name:        reportStringValue(values["Name"]),
		Description: reportStringValue(values["Description"]),
		Kind:        reportStringValue(values["Kind"]),
	}
	if out.Name == "" {
		fieldErrors["Name"] = fmt.Errorf("name is required")
	}
	if _, ok := registry.PairFromPairs(out.Kind, ReportKindChoices); !ok {
		fieldErrors["Kind"] = fmt.Errorf("kind is required")
		return out
	}
	switch out.Kind {
	case "briefing":
		out.BriefingContent = reportStringValue(values["BriefingContent"])
		if out.BriefingContent == "" {
			fieldErrors["BriefingContent"] = fmt.Errorf("briefing content is required")
		}
	case "timeline":
		out.TimelineEntriesJSON = reportStringValue(values["TimelineEntriesJSON"])
		entries, err := parseTimelineEntriesJSON(ctx, out.TimelineEntriesJSON)
		if err != nil {
			fieldErrors["TimelineEntriesJSON"] = err
		} else {
			out.TimelineEntries = entries
		}
	}
	return out
}

func reportStringValue(v any) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func createReportKindRow(tx *gorm.DB, reportID uint, formData parsedReportFormData) error {
	txNoHooks := tx.Session(&gorm.Session{SkipHooks: true})
	switch formData.Kind {
	case "briefing":
		return txNoHooks.Create(&BriefingReport{
			ReportID: reportID,
			Content:  formData.BriefingContent,
		}).Error
	case "timeline":
		timeline := TimelineReport{ReportID: reportID}
		if err := txNoHooks.Create(&timeline).Error; err != nil {
			return err
		}
		entries := make([]TimelineReportEntry, 0, len(formData.TimelineEntries))
		for i, entry := range formData.TimelineEntries {
			entries = append(entries, TimelineReportEntry{
				TimelineReportID: timeline.ID,
				Position:         uint(i),
				Datetime:         entry.Datetime,
				Title:            entry.Title,
				Content:          entry.Content,
			})
		}
		if len(entries) == 0 {
			return nil
		}
		return txNoHooks.Create(&entries).Error
	default:
		return fmt.Errorf("unsupported report kind %q", formData.Kind)
	}
}

func logReportFieldErrors(action string, fieldErrors map[string]error) {
	for field, err := range fieldErrors {
		slog.Error("lacerate: report field error", "action", action, "field", field, "error", err)
	}
}
