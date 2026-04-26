package p_seer_intel

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/views"
)

// intelSourceDetailHrefKey is the request context key for the source app's detail URL
// (empty string when none). Set by [intelSourceDetailHrefLayer] for [seer_intel.IntelDetail] views.
// Read with getters.Key[string](intelSourceDetailHrefKey) in page trees (no DB in getters).
// Key string must be a single path segment (no ".") for top-level [context.Value] storage.
const intelSourceDetailHrefKey = "seer_intel_intelSourceDetailHref"

// intelSourceDetailHrefLayer runs after [views.LayerDetail] for intel; it resolves [LoadIntelKind] +
// [IntelKind.IntelDetail] and stores the path on the context.
type intelSourceDetailHrefLayer struct{}

func (intelSourceDetailHrefLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		in, ok := r.Context().Value("intel").(Intel)
		if !ok {
			ctx := context.WithValue(r.Context(), intelSourceDetailHrefKey, "")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if in.Kind == "" || in.KindID == 0 {
			ctx := context.WithValue(r.Context(), intelSourceDetailHrefKey, "")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db, err := getters.DBFromContext(r.Context())
		if err != nil {
			slog.Error("seer_intel: intel source href layer: db from context", "error", err)
			ctx := context.WithValue(r.Context(), intelSourceDetailHrefKey, "")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		k, err := LoadIntelKind(r.Context(), db, in.Kind, in.KindID)
		if err != nil {
			ctx := context.WithValue(r.Context(), intelSourceDetailHrefKey, "")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		href, err := k.IntelDetail(r.Context())
		if err != nil {
			ctx := context.WithValue(r.Context(), intelSourceDetailHrefKey, "")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx := context.WithValue(r.Context(), intelSourceDetailHrefKey, href)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
