package p_nirmancampus_academicrecords

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type formPatcherAcademicRecordProgramStructureUnitRequired struct{}

func (formPatcherAcademicRecordProgramStructureUnitRequired) Patch(_ views.View, _ *http.Request, values map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if formErrors == nil {
		formErrors = map[string]error{}
	}
	unitID, ok := values["ProgramStructureUnitID"].(uint)
	if !ok || unitID == 0 {
		formErrors["ProgramStructureUnitID"] = fmt.Errorf("select a term")
		return values, formErrors
	}
	return values, formErrors
}

// formPatcherAcademicRecordCreate sets Status default ("Not Applied") and CompulsoryCourses
// from the selected ProgramStructureUnit.
type formPatcherAcademicRecordCreate struct{}

func (formPatcherAcademicRecordCreate) Patch(_ views.View, r *http.Request, values map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if s, ok := values["Status"].(string); !ok || s == "" {
		values["Status"] = "Not Applied"
	}

	tz, _ := r.Context().Value("$tz").(*time.Location)
	if tz == nil {
		tz = components.DefaultTimeZone
	}
	if d, ok := values["Date"].(time.Time); !ok || d.IsZero() {
		now := time.Now().In(tz)
		values["Date"] = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz)
	}

	db, dberr := getters.DBFromContext(r.Context())
	if dberr != nil {
		slog.Error("formPatcherAcademicRecordCreateFromProgramStructure: db from context", "error", dberr)
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
	unitID, okUnit := values["ProgramStructureUnitID"].(uint)
	if !okUnit || unitID == 0 {
		formErrors["ProgramStructureUnitID"] = fmt.Errorf("select a term")
		return values, formErrors
	}

	psu, err := gorm.G[p_nirmancampus_programs.ProgramStructureUnit](db).
		Where("program_id = ? AND id = ?", programID, unitID).
		Preload("CompulsoryCourses", nil).
		Preload("OptionalCourseSelectionPool", nil).
		First(r.Context())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Warn("academic record create: no program structure unit for program",
				"program_id", programID, "program_structure_unit_id", unitID)
		} else {
			slog.Error("academic record create: load program structure unit", "error", err)
		}
		formErrors["ProgramStructureUnitID"] = fmt.Errorf("select a valid term for this program")
		values["CompulsoryCourses"] = components.AssociationIDs{Field: "CompulsoryCourses", IDs: nil}
		return values, formErrors
	}

	ids := make([]uint, 0, len(psu.CompulsoryCourses))
	for _, c := range psu.CompulsoryCourses {
		ids = append(ids, c.ID)
	}
	values["CompulsoryCourses"] = components.AssociationIDs{Field: "CompulsoryCourses", IDs: ids}
	return validateAcademicRecordOptionalCourses(values, formErrors, psu)
}

// formPatcherAcademicRecordUpdate ensures OptionalCourses length matches
// ProgramStructureUnit.OptionalCourseCount for this record.
type formPatcherAcademicRecordUpdate struct{}

func (formPatcherAcademicRecordUpdate) Patch(_ views.View, r *http.Request, values map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if formErrors == nil {
		formErrors = map[string]error{}
	}
	db, dberr := getters.DBFromContext(r.Context())
	if dberr != nil {
		return values, formErrors
	}
	rec, ok := r.Context().Value("academicrecord").(AcademicRecord)
	if !ok || rec.ID == 0 {
		return values, formErrors
	}
	if d, ok := values["Date"].(time.Time); !ok || d.IsZero() {
		values["Date"] = rec.Date
	}
	psu, err := gorm.G[p_nirmancampus_programs.ProgramStructureUnit](db).
		Preload("OptionalCourseSelectionPool", nil).
		Where("id = ? AND program_id = ?", rec.ProgramStructureUnitID, rec.ProgramID).
		First(r.Context())
	if err != nil {
		return values, formErrors
	}
	return validateAcademicRecordOptionalCourses(values, formErrors, psu)
}

func validateAcademicRecordOptionalCourses(values map[string]any, formErrors map[string]error, psu p_nirmancampus_programs.ProgramStructureUnit) (map[string]any, map[string]error) {
	if formErrors == nil {
		formErrors = map[string]error{}
	}
	raw, ok := values["OptionalCourses"]
	if !ok {
		raw = components.AssociationIDs{Field: "OptionalCourses", IDs: nil}
	}
	aids, ok := raw.(components.AssociationIDs)
	if !ok {
		return values, formErrors
	}
	if uint(len(aids.IDs)) != psu.OptionalCourseCount {
		formErrors["OptionalCourses"] = fmt.Errorf("select exactly %d optional course(s) for this program term", psu.OptionalCourseCount)
		return values, formErrors
	}

	allowed := map[uint]struct{}{}
	for _, course := range psu.OptionalCourseSelectionPool {
		allowed[course.ID] = struct{}{}
	}
	for _, id := range aids.IDs {
		if _, ok := allowed[id]; !ok {
			formErrors["OptionalCourses"] = fmt.Errorf("select optional courses from the selected program pool")
			return values, formErrors
		}
	}
	return values, formErrors
}
