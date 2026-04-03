package p_nirmancampus_students

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type studentScopeByRole struct{}

// StudentScopeByRole restricts student queries:
//   - superuser, admin: full queryset
//   - student: only the row linked to the current user (user_id)
//   - any other role: empty queryset
func (studentScopeByRole) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Student]) gorm.ChainInterface[Student] {
	ctx := r.Context()

	rawUser := ctx.Value("$user")
	if rawUser == nil {
		slog.Error("StudentScopeByRole: missing $user in context – auth middleware not applied?")
		panic("StudentScopeByRole: $user is nil in context")
	}
	user, ok := rawUser.(p_users.User)
	if !ok {
		slog.Error("StudentScopeByRole: $user has unexpected type",
			"type", fmt.Sprintf("%T", rawUser),
		)
		panic("StudentScopeByRole: $user has wrong type in context")
	}

	rawRole := ctx.Value("$role")
	if rawRole == nil {
		slog.Error("StudentScopeByRole: missing $role in context – auth middleware not applied?")
		panic("StudentScopeByRole: $role is nil in context")
	}
	roleName, ok := rawRole.(string)
	if !ok {
		slog.Error("StudentScopeByRole: $role has unexpected type",
			"type", fmt.Sprintf("%T", rawRole),
		)
		panic("StudentScopeByRole: $role has wrong type in context")
	}

	switch roleName {
	case "superuser", "admin":
		return query
	case "student":
		return query.Where("user_id = ?", user.ID)
	default:
		return query.Where("1 = 0")
	}
}

var StudentScopeByRole views.QueryPatcher[Student] = studentScopeByRole{}
