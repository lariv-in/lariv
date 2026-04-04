package p_nirmancampus_studentapplications

import (
	"fmt"
	"log/slog"
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

	rawUser := ctx.Value("$user")
	if rawUser == nil {
		slog.Error("StudentApplicationScopeByRole: missing $user in context – auth layer not applied?")
		panic("StudentApplicationScopeByRole: $user is nil in context")
	}
	user, ok := rawUser.(p_users.User)
	if !ok {
		slog.Error("StudentApplicationScopeByRole: $user has unexpected type",
			"type", fmt.Sprintf("%T", rawUser),
		)
		panic("StudentApplicationScopeByRole: $user has wrong type in context")
	}

	rawRole := ctx.Value("$role")
	if rawRole == nil {
		slog.Error("StudentApplicationScopeByRole: missing $role in context – auth layer not applied?")
		panic("StudentApplicationScopeByRole: $role is nil in context")
	}
	roleName, ok := rawRole.(string)
	if !ok {
		slog.Error("StudentApplicationScopeByRole: $role has unexpected type",
			"type", fmt.Sprintf("%T", rawRole),
		)
		panic("StudentApplicationScopeByRole: $role has wrong type in context")
	}

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
