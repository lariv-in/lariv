package p_nirmancampus_academicrecords

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// formPatcherAcademicRecordCreateFromProgramStructure sets Status (default) and CompulsoryCourses
// from the program's ProgramStructureUnit for the submitted Term (TermNumber).
func formPatcherAcademicRecordCreateFromProgramStructure(_ *views.View, r *http.Request, values map[string]any) map[string]any {
	dbVal := r.Context().Value("$db")
	db, ok := dbVal.(*gorm.DB)
	if !ok || db == nil {
		slog.Error("formPatcherAcademicRecordCreateFromProgramStructure: missing $db in context")
		return values
	}

	programID, okPID := uintFromFormValue(values["ProgramID"])
	term, okTerm := intFromFormValue(values["Term"])
	if !okPID || !okTerm {
		return values
	}

	if s, ok := values["Status"].(string); !ok || s == "" {
		values["Status"] = AcademicRecordStatusEnrolled
	}

	var psu p_nirmancampus_programs.ProgramStructureUnit
	err := db.Where("program_id = ? AND term_number = ?", programID, term).
		Preload("CompulsoryCourses").
		First(&psu).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Warn("academic record create: no program structure unit for program/term",
				"program_id", programID, "term", term)
		} else {
			slog.Error("academic record create: load program structure unit", "error", err)
		}
		values["CompulsoryCourses"] = components.AssociationIDs{Field: "CompulsoryCourses", IDs: nil}
		return values
	}

	ids := make([]uint, 0, len(psu.CompulsoryCourses))
	for _, c := range psu.CompulsoryCourses {
		ids = append(ids, c.ID)
	}
	values["CompulsoryCourses"] = components.AssociationIDs{Field: "CompulsoryCourses", IDs: ids}
	return values
}

func uintFromFormValue(v any) (uint, bool) {
	switch x := v.(type) {
	case uint:
		return x, x > 0
	case uint32:
		return uint(x), x > 0
	case uint64:
		return uint(x), x > 0
	case int:
		return uint(x), x > 0
	case int32:
		return uint(x), x > 0
	case int64:
		return uint(x), x > 0
	case float64:
		if x < 0 || x != float64(uint(x)) {
			return 0, false
		}
		u := uint(x)
		return u, u > 0
	default:
		return 0, false
	}
}

func intFromFormValue(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int32:
		return int(x), true
	case int64:
		return int(x), true
	case uint:
		return int(x), true
	case uint32:
		return int(x), true
	case uint64:
		return int(x), true
	case float64:
		if x != float64(int(x)) {
			return 0, false
		}
		return int(x), true
	default:
		return 0, false
	}
}
