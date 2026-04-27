package p_nirmancampus_assignmentsubmissions

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const (
	bulkAddMarksSubmissionsKey   = "assignmentsubmissions.bulk_add_marks_submissions"
	bulkSubmissionMarksFieldName = "BulkSubmissionMarks"
)

type academicRecordBulkMarksForm struct{}

type academicRecordBulkAddMarksLoadLayer struct{}

func (academicRecordBulkAddMarksLoadLayer) Next(view views.View, next http.Handler) http.Handler {
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
		var subs []AssignmentSubmission
		if err := db.Model(&AssignmentSubmission{}).
			Preload("Course").
			Where("academic_record_id = ?", rec.ID).
			Order("id ASC").
			Find(&subs).Error; err != nil {
			next.ServeHTTP(w, r.WithContext(views.ContextWithErrorsAndValues(r.Context(), nil, map[string]error{"_form": err})))
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, bulkAcademicRecordContextKey, rec)
		ctx = context.WithValue(ctx, bulkAddMarksSubmissionsKey, subs)
		ctx = views.ContextWithErrorsAndValues(ctx, map[string]any{"AcademicRecordID": rec.ID, "AcademicRecord": rec}, nil)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bulkAddMarksFromAcademicRecordPostHandler(view *views.View) http.Handler {
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
		byID := bulkAddMarksSubmissionIndex(r.Context())
		if len(byID) == 0 {
			detailURL, err := lago.RoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Static(rec.ID)),
			})(r.Context())
			if err != nil {
				render(map[string]error{"_form": err})
				return
			}
			views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
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

		rows, rowErr := parseBulkSubmissionMarksValue(values[bulkSubmissionMarksFieldName])
		if rowErr != nil {
			errs[bulkSubmissionMarksFieldName] = rowErr
		}

		if len(errs) > 0 {
			render(errs)
			return
		}

		seenID := make(map[uint]struct{}, len(rows))
		for _, row := range rows {
			if row.ID == 0 {
				errs[bulkSubmissionMarksFieldName] = fmt.Errorf("invalid submission id in payload")
				break
			}
			if _, dup := seenID[row.ID]; dup {
				errs[bulkSubmissionMarksFieldName] = fmt.Errorf("duplicate submission id in payload")
				break
			}
			seenID[row.ID] = struct{}{}
			sub, known := byID[row.ID]
			if !known {
				errs[bulkSubmissionMarksFieldName] = fmt.Errorf("one or more submissions are not on this academic record")
				break
			}
			if row.Marks < 0 {
				errs[bulkSubmissionMarksFieldName] = fmt.Errorf("marks cannot be negative")
				break
			}
			if sub.MaxMarks > 0 && row.Marks > sub.MaxMarks {
				errs[bulkSubmissionMarksFieldName] = fmt.Errorf("marks cannot exceed max marks for %q", sub.AssignmentTitle)
				break
			}
		}
		if len(errs) == 0 && len(rows) != len(byID) {
			errs[bulkSubmissionMarksFieldName] = fmt.Errorf("enter marks for every listed submission")
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
		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, row := range rows {
				if err := tx.Model(&AssignmentSubmission{}).
					Where("id = ? AND academic_record_id = ?", row.ID, rec.ID).
					Update("marks", row.Marks).Error; err != nil {
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

func bulkAddMarksSubmissionIndex(ctx context.Context) map[uint]AssignmentSubmission {
	subs, _ := ctx.Value(bulkAddMarksSubmissionsKey).([]AssignmentSubmission)
	out := make(map[uint]AssignmentSubmission, len(subs))
	for _, s := range subs {
		if s.ID != 0 {
			out[s.ID] = s
		}
	}
	return out
}

// bulkSubmissionMarksFormRow is the JSON line item for the bulk marks form hidden field.
type bulkSubmissionMarksFormRow struct {
	ID    uint `json:"id"`
	Marks int  `json:"marks"`
}

func parseBulkSubmissionMarksValue(v any) ([]bulkSubmissionMarksFormRow, error) {
	if v == nil {
		return nil, nil
	}
	switch t := v.(type) {
	case []bulkSubmissionMarksFormRow:
		return t, nil
	default:
		return nil, fmt.Errorf("invalid marks payload type %T", v)
	}
}
