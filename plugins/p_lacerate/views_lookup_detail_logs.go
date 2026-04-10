package p_lacerate

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// ctxKeyLookupLogEntries stores []LookupLogDisplay for the lookup detail page ([lookupDetailLogsLayer]).
const ctxKeyLookupLogEntries = "lookupLogEntries"

// ctxKeyLookupTouchedTargetsOfInterest stores []LookupTouchedTargetOfInterestDisplay ([buildLookupTouchedTargetOfInterestDisplays]).
const ctxKeyLookupTouchedTargetsOfInterest = "lookupTouchedTargetsOfInterest"

// LookupTouchedTargetOfInterestDisplay is a read bundle for the lookup detail UI (not a GORM model).
type LookupTouchedTargetOfInterestDisplay struct {
	TargetOfInterest TargetOfInterest
	Action           string // create | edit
	LogCreatedAt     time.Time
}

func buildLookupTouchedTargetOfInterestDisplays(db *gorm.DB, lookupID uint) []LookupTouchedTargetOfInterestDisplay {
	if lookupID == 0 || db == nil {
		return nil
	}
	var refs []LookupLogTargetOfInterest
	err := db.Joins("INNER JOIN lookup_log_entries ON lookup_log_entries.id = lookup_log_targets_of_interest.lookup_log_entry_id AND lookup_log_entries.deleted_at IS NULL").
		Where("lookup_log_targets_of_interest.lookup_id = ? AND lookup_log_targets_of_interest.deleted_at IS NULL", lookupID).
		Order("lookup_log_entries.created_at DESC").
		Preload("TargetOfInterest").
		Preload("LookupLogEntry").
		Find(&refs).Error
	if err != nil {
		slog.Error("lacerate: lookup touched Targets of Interest", "error", err, "lookup_id", lookupID)
		return nil
	}
	seen := make(map[uint]struct{})
	out := make([]LookupTouchedTargetOfInterestDisplay, 0, len(refs))
	for _, ref := range refs {
		if ref.TargetOfInterestID == 0 || ref.TargetOfInterest.ID == 0 {
			continue
		}
		if _, dup := seen[ref.TargetOfInterestID]; dup {
			continue
		}
		seen[ref.TargetOfInterestID] = struct{}{}
		logAt := time.Time{}
		if ref.LookupLogEntry.ID != 0 {
			logAt = ref.LookupLogEntry.CreatedAt
		}
		out = append(out, LookupTouchedTargetOfInterestDisplay{
			TargetOfInterest: ref.TargetOfInterest,
			Action:           ref.Action,
			LogCreatedAt:     logAt,
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
		touched := buildLookupTouchedTargetOfInterestDisplays(db, lu.ID)
		ctx = context.WithValue(ctx, ctxKeyLookupLogEntries, displays)
		ctx = context.WithValue(ctx, ctxKeyLookupTouchedTargetsOfInterest, touched)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
