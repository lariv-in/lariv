package p_nirmancampus_academicrecords

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const academicRecordProgramStructureUnitContextKey = "academicrecord_program_structure_unit"

// attachAcademicRecordProgramStructureUnitContext loads the ProgramStructureUnit
// for the current AcademicRecord (from the "academicrecord" context key set by
// DetailView) and stores it in context for update-form rendering.
type academicRecordProgramStructureUnitContextMiddleware struct{}

func (academicRecordProgramStructureUnitContextMiddleware) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record, ok := r.Context().Value("academicrecord").(AcademicRecord)
		if !ok || record.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}

		db, ok := r.Context().Value("$db").(*gorm.DB)
		if !ok || db == nil {
			slog.Error("attachAcademicRecordProgramStructureUnitContext: missing $db in context")
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
