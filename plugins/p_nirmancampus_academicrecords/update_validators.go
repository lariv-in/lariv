package p_nirmancampus_academicrecords

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// formValidatorAcademicRecordOptionalCourseCount ensures OptionalCourses length matches
// ProgramStructureUnit.OptionalCourseCount for this record's program and term.
func formValidatorAcademicRecordOptionalCourseCount(_ *views.View, r *http.Request, values map[string]any) map[string]error {
	out := map[string]error{}
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return out
	}
	dbVal := r.Context().Value("$db")
	db, ok := dbVal.(*gorm.DB)
	if !ok || db == nil {
		return out
	}
	var rec AcademicRecord
	if err := db.First(&rec, id).Error; err != nil {
		return out
	}
	var psu p_nirmancampus_programs.ProgramStructureUnit
	if err := db.Select("optional_course_count").
		Where("program_id = ? AND term_number = ?", rec.ProgramID, rec.Term).
		First(&psu).Error; err != nil {
		return out
	}
	expected := psu.OptionalCourseCount
	got := 0
	if raw, ok := values["OptionalCourses"]; ok {
		if aids, ok := raw.(components.AssociationIDs); ok {
			got = len(aids.IDs)
		}
	}
	if got != expected {
		out["OptionalCourses"] = fmt.Errorf("select exactly %d optional course(s) for this program term", expected)
	}
	return out
}
