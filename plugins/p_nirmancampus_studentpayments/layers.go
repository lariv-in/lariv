package p_nirmancampus_studentpayments

import (
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/views"
)

// paymentCreateQueryDefaultsLayer merges ?StudentID= into $in on GET so the create
// form opened from student detail pre-fills the student.
type paymentCreateQueryDefaultsLayer struct{}

func (paymentCreateQueryDefaultsLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		vals := map[string]any{}
		if sid := r.URL.Query().Get("StudentID"); sid != "" {
			if id64, err := strconv.ParseUint(sid, 10, 32); err == nil && id64 != 0 {
				vals["StudentID"] = uint(id64)
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
