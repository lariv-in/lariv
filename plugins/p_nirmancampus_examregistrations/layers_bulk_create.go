package p_nirmancampus_examregistrations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const (
	bulkAcademicRecordContextKey                 = "examregistrations.bulk_academic_record"
	bulkAcademicRecordCoursesWithRegistrationKey = "examregistrations.bulk_courses_with_registration"
)

type academicRecordBulkRegistrationsForm struct{}

const bulkSelectedCourseIDsFieldName = "BulkSelectedCourseIDs"

type academicRecordBulkCreateLoadLayer struct{}

func (academicRecordBulkCreateLoadLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.URL.Query().Get("AcademicRecordID")
		if raw == "" {
			next.ServeHTTP(w, r)
			return
		}
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			next.ServeHTTP(w, r.WithContext(views.ContextWithErrorsAndValues(r.Context(), nil, map[string]error{
				"_form": fmt.Errorf("invalid academic record id"),
			})))
			return
		}
		db, err := getters.DBFromContext(r.Context())
		if err != nil {
			next.ServeHTTP(w, r.WithContext(views.ContextWithErrorsAndValues(r.Context(), nil, map[string]error{"_form": err})))
			return
		}
		q := p_nirmancampus_academicrecords.AcademicRecordQueryPatchersBulkModal.Apply(
			view, r, gorm.G[p_nirmancampus_academicrecords.AcademicRecord](db).Scopes())
		rec, err := q.Where("id = ?", uint(id)).First(r.Context())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = fmt.Errorf("academic record not found")
			}
			next.ServeHTTP(w, r.WithContext(views.ContextWithErrorsAndValues(r.Context(), nil, map[string]error{"_form": err})))
			return
		}
		var courseIDs []uint
		if err := db.Model(&ExamRegistration{}).Where("academic_record_id = ?", rec.ID).Pluck("course_id", &courseIDs).Error; err != nil {
			next.ServeHTTP(w, r.WithContext(views.ContextWithErrorsAndValues(r.Context(), nil, map[string]error{"_form": err})))
			return
		}
		existing := make(map[uint]struct{}, len(courseIDs))
		for _, cid := range courseIDs {
			if cid != 0 {
				existing[cid] = struct{}{}
			}
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, bulkAcademicRecordContextKey, rec)
		ctx = context.WithValue(ctx, bulkAcademicRecordCoursesWithRegistrationKey, existing)
		ctx = views.ContextWithErrorsAndValues(ctx, map[string]any{"AcademicRecordID": rec.ID, "AcademicRecord": rec}, nil)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func allowedCourseIDsForBulk(rec p_nirmancampus_academicrecords.AcademicRecord, alreadyRegistered map[uint]struct{}) map[uint]p_nirmancampus_courses.Course {
	out := make(map[uint]p_nirmancampus_courses.Course)
	add := func(c p_nirmancampus_courses.Course) {
		if c.ID == 0 {
			return
		}
		if alreadyRegistered != nil {
			if _, skip := alreadyRegistered[c.ID]; skip {
				return
			}
		}
		out[c.ID] = c
	}
	for _, c := range rec.CompulsoryCourses {
		add(c)
	}
	for _, c := range rec.OptionalCourses {
		add(c)
	}
	return out
}

func bulkCreateFromAcademicRecordPostHandler(view *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			return
		}
		render := func(errs map[string]error) {
			view.RenderPage(w, r.WithContext(views.ContextWithErrorsAndValues(r.Context(), values, errs)))
		}

		errs := make(map[string]error)
		for k, e := range fieldErrors {
			if e != nil {
				errs[k] = e
			}
		}
		if len(errs) > 0 {
			render(errs)
			return
		}

		rec, ok := r.Context().Value(bulkAcademicRecordContextKey).(p_nirmancampus_academicrecords.AcademicRecord)
		if !ok || rec.ID == 0 {
			render(map[string]error{"_form": fmt.Errorf("academic record not loaded; reopen the form")})
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
		selected, selOK := values[bulkSelectedCourseIDsFieldName].([]uint)
		switch {
		case !selOK:
			errs[bulkSelectedCourseIDsFieldName] = fmt.Errorf("invalid course selection")
		case len(selected) == 0:
			errs[bulkSelectedCourseIDsFieldName] = fmt.Errorf("select at least one course")
		}

		existingRegistered, _ := r.Context().Value(bulkAcademicRecordCoursesWithRegistrationKey).(map[uint]struct{})
		allowed := allowedCourseIDsForBulk(rec, existingRegistered)
		for _, cid := range selected {
			if _, ok := allowed[cid]; !ok {
				errs[bulkSelectedCourseIDsFieldName] = fmt.Errorf("one or more selected courses are not on this academic record")
				break
			}
		}
		if len(errs) > 0 {
			render(errs)
			return
		}

		db, err := getters.DBFromContext(r.Context())
		if err != nil {
			render(map[string]error{"_form": err})
			return
		}
		var n int64
		if err := db.Model(&ExamRegistration{}).Where("academic_record_id = ? AND course_id IN ?", rec.ID, selected).Count(&n).Error; err != nil {
			render(map[string]error{"_form": err})
			return
		}
		if n > 0 {
			render(map[string]error{"_form": fmt.Errorf("one or more selected courses already have an exam registration for this academic record; adjust selection")})
			return
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, cid := range selected {
				c := allowed[cid]
				if err := tx.Create(&ExamRegistration{
					ExamTitle:          c.Name,
					MaxMarks:           0,
					Marks:              0,
					Fee:                c.Fee,
					RegistrationStatus: ExamRegistrationStatusNotRegisteredKey,
					CourseID:           cid,
					AcademicRecordID:   rec.ID,
				}).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			render(map[string]error{"_form": err})
			return
		}
		detailURL, err := lago.RoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(rec.ID)),
		})(r.Context())
		if err != nil {
			render(map[string]error{"_form": err})
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}
