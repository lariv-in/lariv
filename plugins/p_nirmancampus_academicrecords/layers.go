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

// attachAcademicRecordProgramStructureUnitContext loads the ProgramStructureUnit
// for the current AcademicRecord (from the "academicrecord" context key set by
// DetailView) and stores it in context for update-form rendering.
type academicRecordProgramStructureUnitContextLayer struct{}

func (academicRecordProgramStructureUnitContextLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record, ok := r.Context().Value("academicrecord").(AcademicRecord)
		if !ok || record.ID == 0 {
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
			Where("program_id = ? AND term_number = ?", record.ProgramID, record.Term).
			Preload("OptionalCourseSelectionPool").
			First(&psu).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("attachAcademicRecordProgramStructureUnitContext: query failed",
					"error", err,
					"academic_record_id", record.ID,
					"program_id", record.ProgramID,
					"term", record.Term)
			}
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), academicRecordProgramStructureUnitContextKey, psu)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
