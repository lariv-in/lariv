package p_assessments

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

// syllabusTopicMultiSelectURL scopes topic picker by assessment CourseID when set (?CourseID=).
func syllabusTopicMultiSelectURL() getters.Getter[string] {
	base := lago.RoutePath("syllabus.MultiSelectRoute", nil)
	courseID := getters.Key[*uint]("$in.CourseID")
	return func(ctx context.Context) (string, error) {
		s, err := base(ctx)
		if err != nil {
			return "", err
		}
		id, err := courseID(ctx)
		if err != nil {
			return "", err
		}
		if id == nil || *id == 0 {
			return s, nil
		}
		return fmt.Sprintf("%s?CourseID=%d", s, *id), nil
	}
}

// assessmentExamFormStageURLGetter posts multistep progress to the current exam form URL (create or edit).
func assessmentExamFormStageURLGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		r, ok := ctx.Value("$request").(*http.Request)
		if !ok || r == nil || r.URL == nil {
			return lago.RoutePath("assessments.ExamCreateRoute", nil)(ctx)
		}
		return r.URL.RequestURI(), nil
	}
}
