package p_nirmancampus_studentapplications

import (
	"net/http"

	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// Role name from p_users.Role{Name: "unassigned"} (see dashboard_patch.go).
const roleNameUnassigned = "unassigned"

type studentApplicationScopeByRole struct{}

// StudentApplicationScopeByRole restricts application queries:
//   - superuser, admin: full queryset
//   - Unassigned: rows created by the current user (created_by_id)
//   - student and any other role: empty queryset (routes should reject student via layer)
func (studentApplicationScopeByRole) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[StudentApplication]) gorm.ChainInterface[StudentApplication] {
	ctx := r.Context()
	user, roleName := p_users.UserAndRoleFromContext(ctx, "StudentApplicationScopeByRole")

	switch roleName {
	case "superuser", "admin":
		return query
	case roleNameUnassigned:
		return query.Where("created_by_id = ?", user.ID)
	default:
		return query.Where("1 = 0")
	}
}

var StudentApplicationScopeByRole views.QueryPatcher[StudentApplication] = studentApplicationScopeByRole{}
