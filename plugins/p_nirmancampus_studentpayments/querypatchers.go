package p_nirmancampus_studentpayments

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type paymentScopeByRole struct{}

// PaymentScopeByRole restricts payment queries:
// - superuser, admin: full queryset
// - student: only payments for Student rows whose Email matches the logged-in user's email
// - default: empty queryset
func (paymentScopeByRole) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Payment]) gorm.ChainInterface[Payment] {
	ctx := r.Context()
	user, roleName := p_users.UserAndRoleFromContext(ctx, "PaymentScopeByRole")

	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("PaymentScopeByRole: db from context", "error", err)
		panic("PaymentScopeByRole: " + err.Error())
	}

	switch roleName {
	case "superuser":
		return query
	case "admin":
		return query
	case "student":
		email := strings.TrimSpace(user.Email)
		if email == "" {
			return query.Where("1 = 0")
		}
		sub := db.Model(&p_nirmancampus_students.Student{}).Select("id").Where("email = ?", email)
		return query.Where("student_id IN (?)", sub)
	default:
		return query.Where("1 = 0")
	}
}

// PaymentScopeByRole is the default tenant/role scope for payment list/detail/mutations.
var PaymentScopeByRole views.QueryPatcher[Payment] = paymentScopeByRole{}

type paymentListOrder struct{}

func (paymentListOrder) Patch(_ views.View, _ *http.Request, query gorm.ChainInterface[Payment]) gorm.ChainInterface[Payment] {
	return query.Order(`"paid_at" DESC NULLS LAST`).Order("id DESC")
}

// PaymentListOrder sorts list views by paid date then id.
var PaymentListOrder views.QueryPatcher[Payment] = paymentListOrder{}
