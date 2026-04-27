package p_nirmancampus_students

import (
	"net/http"
	"strings"

	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type studentScopeByRole struct{}

// StudentScopeByRole restricts student queries:
//   - superuser, admin: full queryset
//   - student: rows whose Email matches the logged-in user's email (trimmed, non-empty)
//   - any other role: empty queryset
func (studentScopeByRole) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Student]) gorm.ChainInterface[Student] {
	ctx := r.Context()
	user, roleName := p_users.UserAndRoleFromContext(ctx, "StudentScopeByRole")

	switch roleName {
	case "superuser", "admin":
		return query
	case "student":
		email := strings.TrimSpace(user.Email)
		if email == "" {
			return query.Where("1 = 0")
		}
		return query.Where("email = ?", email)
	default:
		return query.Where("1 = 0")
	}
}

var StudentScopeByRole views.QueryPatcher[Student] = studentScopeByRole{}

type studentListOrder struct{}

func (studentListOrder) Patch(_ views.View, _ *http.Request, query gorm.ChainInterface[Student]) gorm.ChainInterface[Student] {
	return query.Order("created_at DESC").Order("id DESC")
}

// StudentListOrder is the default sort for list and select views (newest first).
var StudentListOrder views.QueryPatcher[Student] = studentListOrder{}
