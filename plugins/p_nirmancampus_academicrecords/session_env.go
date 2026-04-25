package p_nirmancampus_academicrecords

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	sessions "github.com/lariv-in/lago/plugins/p_nirmancampus_sessions"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// academicRecordsEnvironmentSessionKey is the $environment cookie map key for the
// academic records list session selector (distinct from TotSchool tally's "session").
const academicRecordsEnvironmentSessionKey = "academicrecords_session"

// AcademicSessionsListGetter returns session id / display label pairs for Environment[uint].
func AcademicSessionsListGetter(ctx context.Context) ([]registry.Pair[uint, string], error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return []registry.Pair[uint, string]{}, nil
	}
	rows, err := gorm.G[sessions.AdmissionSession](db).Order(`"start" DESC`).Find(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]registry.Pair[uint, string], 0, len(rows))
	for _, s := range rows {
		out = append(out, registry.Pair[uint, string]{Key: s.ID, Value: s.Name})
	}
	return out, nil
}

// academicRecordsSessionEnvironmentDefault picks the active session, or the most recent by start.
func academicRecordsSessionEnvironmentDefault(ctx context.Context) (uint, error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return 0, nil
	}
	id, err := defaultAcademicSessionID(ctx, db)
	if err != nil {
		return 0, nil
	}
	return id, nil
}

func defaultAcademicSessionID(ctx context.Context, db *gorm.DB) (uint, error) {
	var active sessions.AdmissionSession
	err := db.Where("is_active = ?", true).Order(`"start" DESC`).First(&active).Error
	if err == nil && active.ID != 0 {
		return active.ID, nil
	}
	var latest sessions.AdmissionSession
	err = db.Order(`"start" DESC`).First(&latest).Error
	if err != nil {
		return 0, err
	}
	return latest.ID, nil
}

// selectedAcademicRecordSessionFilter returns whether to restrict the query to a single
// session id, and that id when restrict is true. When restrict is false, the list shows
// all sessions (user chose "—" in the environment selector).
func selectedAcademicRecordSessionFilter(db *gorm.DB, ctx context.Context) (id uint, restrict bool) {
	envMap, ok := ctx.Value("$environment").(map[string]string)
	if !ok {
		id, err := defaultAcademicSessionID(ctx, db)
		if err != nil {
			slog.Error("selectedAcademicRecordSessionFilter: no default session",
				"error", err,
			)
			return 0, false
		}
		return id, true
	}
	raw, inMap := envMap[academicRecordsEnvironmentSessionKey]
	if !inMap {
		id, err := defaultAcademicSessionID(ctx, db)
		if err != nil {
			slog.Error("selectedAcademicRecordSessionFilter: no default session",
				"error", err,
			)
			return 0, false
		}
		return id, true
	}
	if strings.TrimSpace(raw) == "" {
		return 0, false
	}
	parsed, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || parsed == 0 {
		slog.Error("selectedAcademicRecordSessionFilter: invalid session id in environment",
			"raw", raw,
			"error", err,
		)
		id, defErr := defaultAcademicSessionID(ctx, db)
		if defErr != nil {
			return 0, false
		}
		return id, true
	}
	return uint(parsed), true
}
