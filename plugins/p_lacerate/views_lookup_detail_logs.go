package p_lacerate

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// ctxKeyLookupLogEntries stores components.ObjectList[LookupLogDisplay] for the lookup detail page ([lookupDetailLogsLayer]).
const ctxKeyLookupLogEntries = "lookupLogEntries"

// ctxKeyLookupTouchedReports stores []LookupTouchedReportDisplay from tool_call rows ([buildLookupTouchedReportDisplays]).
const ctxKeyLookupTouchedReports = "lookupTouchedReports"

// LookupTouchedReportDisplay is a read bundle for the lookup detail UI (not a GORM model).
type LookupTouchedReportDisplay struct {
	Report       Report
	Action       string // create | edit
	LogCreatedAt time.Time
}

// lookupLogToolTouchReportLabel maps create/edit report tool names to the short label in the lookup detail sidebar.
var lookupLogToolTouchReportLabel = map[string]string{
	"create_report": "create",
	"edit_report":   "edit",
	// Legacy tool names (pre-rename); safe to remove once no old lookup logs remain.
	"create_target_of_interest": "create",
	"edit_target_of_interest":   "edit",
}

type lookupReportTouchPending struct {
	reportID uint
	action   string
	logAt    time.Time
}

// buildLookupTouchedReportDisplays derives reports touched by this lookup from
// successful create/edit tool_call rows (first occurrence per report ID, log order is newest-first).
func buildLookupTouchedReportDisplays(db *gorm.DB, displays []LookupLogDisplay) []LookupTouchedReportDisplay {
	if db == nil {
		return nil
	}
	var pending []lookupReportTouchPending
	seen := make(map[uint]struct{})
	for _, d := range displays {
		tc := d.ToolCall
		if tc == nil || len(tc.Result) == 0 {
			continue
		}
		label, ok := lookupLogToolTouchReportLabel[tc.Name]
		if !ok {
			continue
		}
		var res struct {
			ID uint `json:"id"`
		}
		if err := json.Unmarshal(tc.Result, &res); err != nil {
			slog.Error("lacerate: lookup tool call result JSON", "error", err, "lookup_log_entry_id", d.ID)
			continue
		}
		if res.ID == 0 {
			continue
		}
		if _, dup := seen[res.ID]; dup {
			continue
		}
		seen[res.ID] = struct{}{}
		pending = append(pending, lookupReportTouchPending{reportID: res.ID, action: label, logAt: d.CreatedAt})
	}
	if len(pending) == 0 {
		return nil
	}
	ids := make([]uint, len(pending))
	for i := range pending {
		ids[i] = pending[i].reportID
	}
	var reports []Report
	if err := db.Where("id IN ?", ids).Find(&reports).Error; err != nil {
		slog.Error("lacerate: lookup touched reports load", "error", err)
		return nil
	}
	byID := make(map[uint]Report, len(reports))
	for _, t := range reports {
		byID[t.ID] = t
	}
	out := make([]LookupTouchedReportDisplay, 0, len(pending))
	for _, p := range pending {
		t, ok := byID[p.reportID]
		if !ok {
			continue
		}
		out = append(out, LookupTouchedReportDisplay{
			Report:       t,
			Action:       p.action,
			LogCreatedAt: p.logAt,
		})
	}
	return out
}

// LookupLogDisplay bundles a persisted [LookupLogEntry] with its payload row for UI only (not a GORM model).
type LookupLogDisplay struct {
	LookupLogEntry
	Thought   *LookupThought
	LogText   *LookupText
	ToolCall  *LookupToolCall
	ToolError *LookupToolError
}

// lookupLogFirstPayload loads the child row for one log entry kind; returns nil on missing row or error (after logging non-NotFound).
func lookupLogFirstPayload[T any](db *gorm.DB, entryID uint, kindSlug string) *T {
	var row T
	err := db.Where("lookup_log_entry_id = ?", entryID).First(&row).Error
	if err == nil {
		return &row
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("lacerate: lookup log "+kindSlug+" payload", "error", err, "lookup_log_entry_id", entryID)
	}
	return nil
}

type lookupLogPayloadAttach func(db *gorm.DB, e LookupLogEntry, d *LookupLogDisplay)

// lookupLogDisplayPayloadAttachers maps [LookupLogEntry.Kind] to how the corresponding payload row is attached (one query per row when present).
var lookupLogDisplayPayloadAttachers = map[string]lookupLogPayloadAttach{
	"thought": func(db *gorm.DB, e LookupLogEntry, d *LookupLogDisplay) {
		d.Thought = lookupLogFirstPayload[LookupThought](db, e.ID, "thought")
	},
	"text": func(db *gorm.DB, e LookupLogEntry, d *LookupLogDisplay) {
		d.LogText = lookupLogFirstPayload[LookupText](db, e.ID, "text")
	},
	"tool_call": func(db *gorm.DB, e LookupLogEntry, d *LookupLogDisplay) {
		d.ToolCall = lookupLogFirstPayload[LookupToolCall](db, e.ID, "tool_call")
	},
	"tool_error": func(db *gorm.DB, e LookupLogEntry, d *LookupLogDisplay) {
		d.ToolError = lookupLogFirstPayload[LookupToolError](db, e.ID, "tool_error")
	},
}

// lookupLogDisplaysWithPayloads loads child rows by kind for each entry (one query per log row).
func lookupLogDisplaysWithPayloads(db *gorm.DB, entries []LookupLogEntry) []LookupLogDisplay {
	out := make([]LookupLogDisplay, 0, len(entries))
	for _, e := range entries {
		d := LookupLogDisplay{LookupLogEntry: e}
		if fn, ok := lookupLogDisplayPayloadAttachers[e.Kind]; ok {
			fn(db, e, &d)
		}
		out = append(out, d)
	}
	return out
}

type lookupDetailLogsLayer struct{}

func (lookupDetailLogsLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		lu, ok := ctx.Value("lookup").(Lookup)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok || db == nil {
			slog.Error("lacerate: lookup detail logs: missing db in context")
			next.ServeHTTP(w, r)
			return
		}
		var entries []LookupLogEntry
		if err := db.Where("lookup_id = ?", lu.ID).
			Order("created_at DESC").
			Find(&entries).Error; err != nil {
			slog.Error("lacerate: lookup detail logs: query", "error", err, "lookup_id", lu.ID)
			next.ServeHTTP(w, r)
			return
		}
		displays := lookupLogDisplaysWithPayloads(db, entries)
		touched := buildLookupTouchedReportDisplays(db, displays)
		n := uint64(len(displays))
		ctx = context.WithValue(ctx, ctxKeyLookupLogEntries, components.ObjectList[LookupLogDisplay]{
			Items:    displays,
			Number:   1,
			NumPages: 1,
			Total:    n,
		})
		ctx = context.WithValue(ctx, ctxKeyLookupTouchedReports, touched)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
