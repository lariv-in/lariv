package p_users

import (
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/views"
)

// A view layer that only allows superuser authenticated users to continue.
type SuperuserOnlyLayer struct{}

func (SuperuserOnlyLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context(), "finance_taxes.SuperuserOnlyLayer")
		if !user.IsSuperuser {
			slog.Error("finance_taxes.SuperuserOnlyLayer: forbidden", "user_id", user.ID)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
