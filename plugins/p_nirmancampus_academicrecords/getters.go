package p_nirmancampus_academicrecords

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
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
	id, err := sessions.DefaultAdmissionSessionID(db)
	if err != nil {
		return 0, nil
	}
	return id, nil
}

func optionalCourseCountDisplayGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		psu, err := getters.Key[p_nirmancampus_programs.ProgramStructureUnit](academicRecordProgramStructureUnitContextKey)(ctx)
		if err != nil || psu.ID == 0 {
			return "—", nil
		}
		return fmt.Sprintf("%d", psu.OptionalCourseCount), nil
	}
}

func optionalCoursesMultiSelectURLGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		base, err := lago.RoutePath("courses.MultiSelectRoute", nil)(ctx)
		if err != nil {
			return "", err
		}
		u, errParse := url.Parse(base)
		if errParse != nil {
			return base, nil
		}
		psu, err := getters.Key[p_nirmancampus_programs.ProgramStructureUnit](academicRecordProgramStructureUnitContextKey)(ctx)
		q := u.Query()
		if err != nil || psu.ID == 0 || len(psu.OptionalCourseSelectionPool) == 0 {
			q.Set("pool_course_ids", "")
		} else {
			parts := make([]string, 0, len(psu.OptionalCourseSelectionPool))
			for _, c := range psu.OptionalCourseSelectionPool {
				parts = append(parts, strconv.FormatUint(uint64(c.ID), 10))
			}
			q.Set("pool_course_ids", strings.Join(parts, ","))
		}
		u.RawQuery = q.Encode()
		return u.String(), nil
	}
}

func academicRecordCreateStageURLGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		r, ok := ctx.Value("$request").(*http.Request)
		if !ok || r == nil || r.URL == nil {
			return lago.RoutePath("academicrecords.CreateRoute", nil)(ctx)
		}
		return r.URL.RequestURI(), nil
	}
}

func programStructureUnitDisplayGetter() getters.Getter[string] {
	return getters.Format("Term %d", getters.Any(getters.Key[uint]("$in.TermNumber")))
}

func programStructureUnitSelectURLGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		base, err := lago.RoutePath("academicrecords.ProgramStructureUnitSelectRoute", nil)(ctx)
		if err != nil {
			return "", err
		}
		programID, ok := academicRecordProgramIDForChoices(ctx)
		if !ok || programID == 0 {
			return base, nil
		}
		u, err := url.Parse(base)
		if err != nil {
			return base, nil
		}
		q := u.Query()
		q.Set("ProgramID", strconv.FormatUint(uint64(programID), 10))
		u.RawQuery = q.Encode()
		return u.String(), nil
	}
}

func academicRecordProgramIDForChoices(ctx context.Context) (uint, bool) {
	if programID, ok := academicRecordProgramIDFromContext(ctx); ok {
		return programID, true
	}
	req, ok := ctx.Value("$request").(*http.Request)
	if !ok || req == nil {
		return 0, false
	}
	raw := req.FormValue("ProgramID")
	if raw == "" {
		return 0, false
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

func academicRecordDefaultGetter(base getters.Getter[time.Time]) getters.Getter[time.Time] {
	return func(ctx context.Context) (time.Time, error) {
		t, err := base(ctx)
		if err != nil {
			return time.Time{}, err
		}
		if !t.IsZero() {
			return t, nil
		}
		tz, _ := ctx.Value("$tz").(*time.Location)
		if tz == nil {
			tz = components.DefaultTimeZone
		}
		now := time.Now().In(tz)
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz), nil
	}
}
