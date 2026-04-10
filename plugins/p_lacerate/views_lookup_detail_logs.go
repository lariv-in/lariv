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

// ctxKeyLookupTouchedTargetsOfInterest stores []LookupTouchedTargetOfInterestDisplay from tool_call rows ([buildLookupTouchedTargetOfInterestDisplays]).
const ctxKeyLookupTouchedTargetsOfInterest = "lookupTouchedTargetsOfInterest"

// LookupTouchedTargetOfInterestDisplay is a read bundle for the lookup detail UI (not a GORM model).
type LookupTouchedTargetOfInterestDisplay struct {
	TargetOfInterest TargetOfInterest
	Action           string // create | edit
	LogCreatedAt     time.Time
}

func buildLookupTouchedTargetOfInterestDisplays(db *gorm.DB, displays []LookupLogDisplay) []LookupTouchedTargetOfInterestDisplay {
	if db == nil {
		return nil
	}
	type pending struct {
		id     uint
		action string
		logAt  time.Time
	}
	var pend []pending
	seen := make(map[uint]struct{})
	for _, d := range displays {
		tc := d.ToolCall
		if tc == nil || len(tc.Result) == 0 {
			continue
		}
		var action string
		switch tc.Name {
		case "create_target_of_interest":
			action = "create"
		case "edit_target_of_interest":
			action = "edit"
		default:
			continue
		}
		var toolOut struct {
			ID uint `json:"id"`
		}
		if err := json.Unmarshal(tc.Result, &toolOut); err != nil {
			slog.Error("lacerate: lookup tool call result JSON", "error", err, "lookup_log_entry_id", d.ID)
			continue
		}
		id := toolOut.ID
		if id == 0 {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		pend = append(pend, pending{id: id, action: action, logAt: d.CreatedAt})
	}
	if len(pend) == 0 {
		return nil
	}
	ids := make([]uint, len(pend))
	for i, p := range pend {
		ids[i] = p.id
	}
	var targets []TargetOfInterest
	if err := db.Where("id IN ?", ids).Find(&targets).Error; err != nil {
		slog.Error("lacerate: lookup touched Targets of Interest load", "error", err)
		return nil
	}
	byID := make(map[uint]TargetOfInterest, len(targets))
	for _, t := range targets {
		byID[t.ID] = t
	}
	out := make([]LookupTouchedTargetOfInterestDisplay, 0, len(pend))
	for _, p := range pend {
		t, ok := byID[p.id]
		if !ok || t.ID == 0 {
			continue
		}
		out = append(out, LookupTouchedTargetOfInterestDisplay{
			TargetOfInterest: t,
			Action:           p.action,
			LogCreatedAt:     p.logAt,
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

// lookupLogDisplaysWithPayloads loads child rows by kind for each entry (one query per log row).
func lookupLogDisplaysWithPayloads(db *gorm.DB, entries []LookupLogEntry) []LookupLogDisplay {
	out := make([]LookupLogDisplay, 0, len(entries))
	for _, e := range entries {
		d := LookupLogDisplay{LookupLogEntry: e}
		switch e.Kind {
		case "thought":
			var th LookupThought
			if err := db.Where("lookup_log_entry_id = ?", e.ID).First(&th).Error; err == nil {
				d.Thought = &th
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("lacerate: lookup log thought payload", "error", err, "lookup_log_entry_id", e.ID)
			}
		case "text":
			var tx LookupText
			if err := db.Where("lookup_log_entry_id = ?", e.ID).First(&tx).Error; err == nil {
				d.LogText = &tx
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("lacerate: lookup log text payload", "error", err, "lookup_log_entry_id", e.ID)
			}
		case "tool_call":
			var tc LookupToolCall
			if err := db.Where("lookup_log_entry_id = ?", e.ID).First(&tc).Error; err == nil {
				d.ToolCall = &tc
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("lacerate: lookup log tool_call payload", "error", err, "lookup_log_entry_id", e.ID)
			}
		case "tool_error":
			var te LookupToolError
			if err := db.Where("lookup_log_entry_id = ?", e.ID).First(&te).Error; err == nil {
				d.ToolError = &te
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("lacerate: lookup log tool_error payload", "error", err, "lookup_log_entry_id", e.ID)
			}
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
		touched := buildLookupTouchedTargetOfInterestDisplays(db, displays)
		n := uint64(len(displays))
		ctx = context.WithValue(ctx, ctxKeyLookupLogEntries, components.ObjectList[LookupLogDisplay]{
			Items:    displays,
			Number:   1,
			NumPages: 1,
			Total:    n,
		})
		ctx = context.WithValue(ctx, ctxKeyLookupTouchedTargetsOfInterest, touched)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
