package p_nirmancampus_academicrecords

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// formPatcherAcademicRecordCreateFromProgramStructure sets Status (default) and CompulsoryCourses
// from the program's ProgramStructureUnit for the submitted Term (TermNumber).
func formPatcherAcademicRecordCreateFromProgramStructure(_ *views.View, r *http.Request, values map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	dbVal := r.Context().Value("$db")
	db, ok := dbVal.(*gorm.DB)
	if !ok || db == nil {
		slog.Error("formPatcherAcademicRecordCreateFromProgramStructure: missing $db in context")
		return values, formErrors
	}
	if formErrors == nil {
		formErrors = map[string]error{}
	}

	programID, okPID := values["ProgramID"].(uint)
	if !okPID || programID == 0 {
		formErrors["ProgramID"] = fmt.Errorf("select a program")
		return values, formErrors
	}
	term, okTerm := values["Term"].(int)
	if !okTerm {
		formErrors["Term"] = fmt.Errorf("enter a valid term")
		return values, formErrors
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
		return values, formErrors
	}

	ids := make([]uint, 0, len(psu.CompulsoryCourses))
	for _, c := range psu.CompulsoryCourses {
		ids = append(ids, c.ID)
	}
	values["CompulsoryCourses"] = components.AssociationIDs{Field: "CompulsoryCourses", IDs: ids}
	return values, formErrors
}
