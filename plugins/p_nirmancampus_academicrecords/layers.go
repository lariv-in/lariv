package p_nirmancampus_academicrecords

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const academicRecordProgramStructureUnitContextKey = "academicrecord_program_structure_unit"
const academicRecordProgramStructureUnitsContextKey = "academicrecord_program_structure_units"

type academicRecordProgramStructureUnitsContextLayer struct{}

func (academicRecordProgramStructureUnitsContextLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		programID, ok := academicRecordProgramIDFromContext(r.Context())
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("attachAcademicRecordProgramStructureUnitsContext: db from context", "error", dberr)
			next.ServeHTTP(w, r)
			return
		}

		var units []p_nirmancampus_programs.ProgramStructureUnit
		if err := db.
			Where("program_id = ?", programID).
			Order("term_number ASC").
			Find(&units).Error; err != nil {
			slog.Error("attachAcademicRecordProgramStructureUnitsContext: query failed",
				"error", err,
				"program_id", programID)
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), academicRecordProgramStructureUnitsContextKey, units)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// attachAcademicRecordProgramStructureUnitContext loads the ProgramStructureUnit
// for the current AcademicRecord (from the "academicrecord" context key set by
// DetailView) or the current form values in $in and stores it in context for
// form rendering.
type academicRecordProgramStructureUnitContextLayer struct{}

func (academicRecordProgramStructureUnitContextLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		programID, unitID, ok := academicRecordProgramAndStructureUnitFromContext(r.Context())
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("attachAcademicRecordProgramStructureUnitContext: db from context", "error", dberr)
			next.ServeHTTP(w, r)
			return
		}

		var psu p_nirmancampus_programs.ProgramStructureUnit
		err := db.
			Where("program_id = ? AND id = ?", programID, unitID).
			Preload("CompulsoryCourses").
			Preload("OptionalCourseSelectionPool").
			First(&psu).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("attachAcademicRecordProgramStructureUnitContext: query failed",
					"error", err,
					"program_id", programID,
					"program_structure_unit_id", unitID)
			}
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), academicRecordProgramStructureUnitContextKey, psu)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func academicRecordProgramIDFromContext(ctx context.Context) (uint, bool) {
	if record, ok := ctx.Value("academicrecord").(AcademicRecord); ok && record.ID != 0 {
		return record.ProgramID, true
	}

	values, ok := ctx.Value(getters.ContextKeyIn).(map[string]any)
	if !ok || values == nil {
		return 0, false
	}
	programID, okProgram := values["ProgramID"].(uint)
	if !okProgram || programID == 0 {
		return 0, false
	}
	return programID, true
}

func academicRecordProgramAndStructureUnitFromContext(ctx context.Context) (uint, uint, bool) {
	if record, ok := ctx.Value("academicrecord").(AcademicRecord); ok && record.ID != 0 {
		return record.ProgramID, record.ProgramStructureUnitID, true
	}

	values, ok := ctx.Value(getters.ContextKeyIn).(map[string]any)
	if !ok || values == nil {
		return 0, 0, false
	}
	programID, okProgram := values["ProgramID"].(uint)
	unitID, okUnit := values["ProgramStructureUnitID"].(uint)
	if !okProgram || !okUnit || programID == 0 || unitID == 0 {
		return 0, 0, false
	}
	return programID, unitID, true
}

// academicRecordCreateQueryDefaultsLayer merges select query params into $in on GET
// so e.g. ?StudentID= from the student detail table pre-fills the create form.
type academicRecordCreateQueryDefaultsLayer struct{}

func (academicRecordCreateQueryDefaultsLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		vals := map[string]any{}
		if sid := r.URL.Query().Get("StudentID"); sid != "" {
			if id64, err := strconv.ParseUint(sid, 10, 32); err == nil && id64 != 0 {
				vals["StudentID"] = uint(id64)
			}
		}
		if r.URL.Query().Get("SessionID") == "" {
			if db, err := getters.DBFromContext(r.Context()); err == nil {
				sessionID, restrict := selectedAcademicRecordSessionFilter(db, r.Context())
				if restrict && sessionID > 0 {
					vals["SessionID"] = sessionID
				}
			}
		}
		if len(vals) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		ctx := views.ContextWithErrorsAndValues(r.Context(), vals, nil)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
