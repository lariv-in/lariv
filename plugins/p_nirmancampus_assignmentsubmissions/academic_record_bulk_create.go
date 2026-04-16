package p_nirmancampus_assignmentsubmissions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// bulkAcademicRecordContextKey holds the loaded academic record (with courses) for the bulk-create modal.
const bulkAcademicRecordContextKey = "assignmentsubmissions.bulk_academic_record"

// academicRecordBulkSubmissionsForm is a marker type for the bulk-create modal form (no DB table).
type academicRecordBulkSubmissionsForm struct{}

const bulkSelectedCourseIDsFieldName = "BulkSelectedCourseIDs"

type academicRecordBulkCreateLoadLayer struct{}

func (academicRecordBulkCreateLoadLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		aidStr := r.URL.Query().Get("AcademicRecordID")
		if aidStr == "" {
			next.ServeHTTP(w, r)
			return
		}
		id64, err := strconv.ParseUint(aidStr, 10, 64)
		if err != nil || id64 == 0 {
			ctx := views.ContextWithErrorsAndValues(r.Context(), nil, map[string]error{
				"_form": fmt.Errorf("invalid academic record id"),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("academicRecordBulkCreateLoadLayer: db from context", "error", dberr)
			ctx := views.ContextWithErrorsAndValues(r.Context(), nil, map[string]error{"_form": dberr})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		var rec p_nirmancampus_academicrecords.AcademicRecord
		if err := db.Preload("CompulsoryCourses").Preload("OptionalCourses").Preload("Student").
			First(&rec, uint(id64)).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx := views.ContextWithErrorsAndValues(r.Context(), nil, map[string]error{
					"_form": fmt.Errorf("academic record not found"),
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			slog.Error("academicRecordBulkCreateLoadLayer: load failed", "error", err, "id", id64)
			ctx := views.ContextWithErrorsAndValues(r.Context(), nil, map[string]error{"_form": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx := context.WithValue(r.Context(), bulkAcademicRecordContextKey, rec)
		ctx = views.ContextWithErrorsAndValues(ctx, map[string]any{"AcademicRecordID": rec.ID}, nil)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func allowedCourseIDsForBulk(rec p_nirmancampus_academicrecords.AcademicRecord) map[uint]p_nirmancampus_courses.Course {
	out := make(map[uint]p_nirmancampus_courses.Course)
	for _, c := range rec.CompulsoryCourses {
		if c.ID != 0 {
			out[c.ID] = c
		}
	}
	for _, c := range rec.OptionalCourses {
		if c.ID != 0 {
			out[c.ID] = c
		}
	}
	return out
}

func bulkCreateFromAcademicRecordPostHandler(view *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			return
		}
		errs := map[string]error{}
		for k, e := range fieldErrors {
			if e != nil {
				errs[k] = e
			}
		}
		if len(errs) > 0 {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, errs)
			view.RenderPage(w, r.WithContext(ctx))
			return
		}
		rec, ok := r.Context().Value(bulkAcademicRecordContextKey).(p_nirmancampus_academicrecords.AcademicRecord)
		if !ok || rec.ID == 0 {
			slog.Error("bulkCreateFromAcademicRecordPostHandler: academic record missing from context")
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{
				"_form": fmt.Errorf("academic record not loaded; reopen the form"),
			})
			view.RenderPage(w, r.WithContext(ctx))
			return
		}

		aidRaw, hasAID := values["AcademicRecordID"]
		if !hasAID {
			errs["AcademicRecordID"] = fmt.Errorf("academic record is required")
		} else if aid, ok := aidRaw.(uint); !ok {
			errs["AcademicRecordID"] = fmt.Errorf("AcademicRecordID: wrong type %T (expected uint)", aidRaw)
		} else if aid != rec.ID {
			errs["AcademicRecordID"] = fmt.Errorf("academic record mismatch")
		}

		rawSel, hasSel := values[bulkSelectedCourseIDsFieldName]
		var selected []uint
		if hasSel {
			var ok bool
			selected, ok = rawSel.([]uint)
			if !ok {
				errs[bulkSelectedCourseIDsFieldName] = fmt.Errorf("invalid course selection type %T", rawSel)
			}
		}
		if len(selected) == 0 && errs[bulkSelectedCourseIDsFieldName] == nil {
			errs[bulkSelectedCourseIDsFieldName] = fmt.Errorf("select at least one course")
		}

		allowed := allowedCourseIDsForBulk(rec)
		for _, cid := range selected {
			if _, ok := allowed[cid]; !ok {
				errs[bulkSelectedCourseIDsFieldName] = fmt.Errorf("one or more selected courses are not on this academic record")
				break
			}
		}

		if len(errs) > 0 {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, errs)
			view.RenderPage(w, r.WithContext(ctx))
			return
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("bulkCreateFromAcademicRecordPostHandler: db from context", "error", dberr)
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": dberr})
			view.RenderPage(w, r.WithContext(ctx))
			return
		}

		var clashCount int64
		if err := db.Model(&AssignmentSubmission{}).
			Where("academic_record_id = ? AND course_id IN ?", rec.ID, selected).
			Count(&clashCount).Error; err != nil {
			slog.Error("bulkCreateFromAcademicRecordPostHandler: duplicate check failed", "error", err)
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": err})
			view.RenderPage(w, r.WithContext(ctx))
			return
		}
		if clashCount > 0 {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{
				"_form": fmt.Errorf("one or more selected courses already have a submission for this academic record; adjust selection"),
			})
			view.RenderPage(w, r.WithContext(ctx))
			return
		}

		txErr := db.Transaction(func(tx *gorm.DB) error {
			for _, cid := range selected {
				course := allowed[cid]
				row := AssignmentSubmission{
					AssignmentTitle:    course.Name,
					MaxMarks:           0,
					Marks:              0,
					SubmissionStatus: AssignmentSubmissionStatusCreatedKey,
					CourseID:           cid,
					AcademicRecordID:   rec.ID,
				}
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if txErr != nil {
			slog.Error("bulkCreateFromAcademicRecordPostHandler: transaction failed", "error", txErr)
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": txErr})
			view.RenderPage(w, r.WithContext(ctx))
			return
		}

		detailURL, urlErr := lago.RoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(rec.ID)),
		})(r.Context())
		if urlErr != nil {
			slog.Error("bulkCreateFromAcademicRecordPostHandler: detail URL", "error", urlErr)
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": urlErr})
			view.RenderPage(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}
