package p_nirmancampus_assignmentsubmissions

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type assignmentSubmissionCreateFormPatcher struct{}

func courseIDFromFormData(formData map[string]any) (uint, error) {
	raw, ok := formData["CourseID"]
	if !ok || raw == nil {
		return 0, fmt.Errorf("missing course")
	}
	courseID, ok := raw.(uint)
	if !ok {
		return 0, fmt.Errorf("CourseID: wrong type %T (expected uint from form inputs)", raw)
	}
	if courseID == 0 {
		return 0, fmt.Errorf("missing course")
	}
	return courseID, nil
}

func (assignmentSubmissionCreateFormPatcher) Patch(_ views.View, r *http.Request, formData map[string]any, _ map[string]error) (map[string]any, map[string]error) {
	courseID, err := courseIDFromFormData(formData)
	if err != nil {
		return nil, map[string]error{"CourseID": err}
	}
	db, dberr := getters.DBFromContext(r.Context())
	if dberr != nil {
		return nil, map[string]error{"_form": dberr}
	}
	var course p_nirmancampus_courses.Course
	if err := db.First(&course, courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, map[string]error{"CourseID": fmt.Errorf("course not found")}
		}
		return nil, map[string]error{"_form": err}
	}
	return map[string]any{
		"AssignmentTitle":    course.Name,
		"SubmissionStatus": AssignmentSubmissionStatusCreatedKey,
		"MaxMarks":           0,
		"Marks":              0,
	}, nil
}
