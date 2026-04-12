package p_nirmancampus_academicrecords

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const academicRecordTermMax = 6

// formPatcherAcademicRecordTermMax rejects Term greater than academicRecordTermMax.
type formPatcherAcademicRecordTermMax struct{}

func (formPatcherAcademicRecordTermMax) Patch(_ views.View, _ *http.Request, values map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if formErrors == nil {
		formErrors = map[string]error{}
	}
	term, ok := values["Term"].(uint)
	if !ok {
		return values, formErrors
	}
	if term > academicRecordTermMax {
		formErrors["Term"] = fmt.Errorf("term must be less than or equal to %d", academicRecordTermMax)
	}
	return values, formErrors
}

// formPatcherAcademicRecordCreate sets Status (default) and CompulsoryCourses
// from the program's ProgramStructureUnit for the submitted Term (TermNumber).
type formPatcherAcademicRecordCreate struct{}

func (formPatcherAcademicRecordCreate) Patch(_ views.View, r *http.Request, values map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if s, ok := values["Status"].(string); !ok || s == "" {
		values["Status"] = "Enrolled"
	}

	tz, _ := r.Context().Value("$tz").(*time.Location)
	if tz == nil {
		tz = components.DefaultTimeZone
	}
	if d, ok := values["Date"].(time.Time); !ok || d.IsZero() {
		now := time.Now().In(tz)
		values["Date"] = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz)
	}

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
	term, okTerm := values["Term"].(uint)
	if !okTerm {
		formErrors["Term"] = fmt.Errorf("enter a valid term")
		return values, formErrors
	}

	psu, err := gorm.G[p_nirmancampus_programs.ProgramStructureUnit](db).
		Where("program_id = ? AND term_number = ?", programID, term).
		Preload("CompulsoryCourses", nil).
		First(r.Context())
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

// formPatcherAcademicRecordUpdate ensures OptionalCourses length matches
// ProgramStructureUnit.OptionalCourseCount for this record's program and term.
type formPatcherAcademicRecordUpdate struct{}

func (formPatcherAcademicRecordUpdate) Patch(_ views.View, r *http.Request, values map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if formErrors == nil {
		formErrors = map[string]error{}
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return values, formErrors
	}
	dbVal := r.Context().Value("$db")
	db, ok := dbVal.(*gorm.DB)
	if !ok || db == nil {
		return values, formErrors
	}
	rec, err := gorm.G[AcademicRecord](db).Where("id = ?", id).First(r.Context())
	if err != nil {
		return values, formErrors
	}
	if d, ok := values["Date"].(time.Time); !ok || d.IsZero() {
		values["Date"] = rec.Date
	}
	psu, err := gorm.G[p_nirmancampus_programs.ProgramStructureUnit](db).
		Select("optional_course_count").
		Where("program_id = ? AND term_number = ?", rec.ProgramID, rec.Term).
		First(r.Context())
	if err != nil {
		return values, formErrors
	}
	expected := psu.OptionalCourseCount
	got := 0
	if raw, ok := values["OptionalCourses"]; ok {
		if aids, ok := raw.(components.AssociationIDs); ok {
			got = len(aids.IDs)
		}
	}
	if uint(got) != expected {
		formErrors["OptionalCourses"] = fmt.Errorf("select exactly %d optional course(s) for this program term", expected)
	}
	return values, formErrors
}
