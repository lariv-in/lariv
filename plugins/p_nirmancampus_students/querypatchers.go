package p_nirmancampus_students

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type userPickForStudentQueryPatcher struct{}

// UserPickForStudentQueryPatcher limits the user picker to accounts with the
// student role that are not already linked to a Student row. Optional query
// param allow_user_id includes that user so edit forms can keep the current link.
func (userPickForStudentQueryPatcher) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[p_users.User]) gorm.ChainInterface[p_users.User] {
	db, ok := r.Context().Value("$db").(*gorm.DB)
	if !ok || db == nil {
		return query
	}
	noStudentSub := db.Model(&Student{}).Select("user_id")
	allowStr := r.URL.Query().Get("allow_user_id")
	if allowStr != "" {
		id, err := strconv.ParseUint(allowStr, 10, 32)
		if err == nil && id > 0 {
			return query.Where(
				"(role_id = (SELECT id FROM roles WHERE name = ? LIMIT 1) AND id NOT IN (?)) OR id = ?",
				"student",
				noStudentSub,
				uint(id),
			)
		}
	}
	return query.Where("role_id = (SELECT id FROM roles WHERE name = ? LIMIT 1) AND id NOT IN (?)", "student", noStudentSub)
}

// UserPickForStudentQueryPatcher is the query patcher for students.UserPickView.
var UserPickForStudentQueryPatcher views.QueryPatcher[p_users.User] = userPickForStudentQueryPatcher{}

type studentScopeByRole struct{}

// StudentScopeByRole restricts student queries:
//   - superuser, admin: full queryset
//   - student: only the row linked to the current user (user_id)
//   - any other role: empty queryset
func (studentScopeByRole) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Student]) gorm.ChainInterface[Student] {
	ctx := r.Context()

	rawUser := ctx.Value("$user")
	if rawUser == nil {
		slog.Error("StudentScopeByRole: missing $user in context – auth layer not applied?")
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
		slog.Error("StudentScopeByRole: missing $role in context – auth layer not applied?")
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
