package p_nirmancampus_examregistrations

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const listFilterAcademicRecordContextKey = "examregistrations.list_filter_academic_record"

type listFilterAcademicRecordLoadLayer struct{}

func (listFilterAcademicRecordLoadLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		raw := r.URL.Query().Get("AcademicRecordID")
		if raw == "" {
			next.ServeHTTP(w, r)
			return
		}
		id64, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id64 == 0 {
			next.ServeHTTP(w, r)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			next.ServeHTTP(w, r)
			return
		}
		q := p_nirmancampus_academicrecords.AcademicRecordQueryPatchersAssignmentSubmissionInput.Apply(
			view, r, gorm.G[p_nirmancampus_academicrecords.AcademicRecord](db).Scopes())
		rec, err := q.Where("id = ?", uint(id64)).First(r.Context())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				next.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), listFilterAcademicRecordContextKey, rec)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
