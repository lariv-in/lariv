package p_nirmancampus_assignmentsubmissions

import (
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/views"
)

// assignmentSubmissionCreateQueryDefaultsLayer merges query params into $in on GET
// (e.g. ?AcademicRecordID= from the academic record detail table pre-fills the create modal).
type assignmentSubmissionCreateQueryDefaultsLayer struct{}

func (assignmentSubmissionCreateQueryDefaultsLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		vals := map[string]any{}
		if aid := r.URL.Query().Get("AcademicRecordID"); aid != "" {
			if id64, err := strconv.ParseUint(aid, 10, 32); err == nil && id64 != 0 {
				vals["AcademicRecordID"] = uint(id64)
			}
		}
		if len(vals) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		ctx := views.ContextWithErrorsAndValues(r.Context(), vals, nil)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
