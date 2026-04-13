package p_users

import (
	"context"
	"fmt"
	"log/slog"
)

// UserFromContext returns $user from ctx. caller labels slog/panic (e.g. "CourseScopeByRole").
func UserFromContext(ctx context.Context, caller string) User {
	user, ok := ctx.Value("$user").(User)
	if !ok {
		slog.Error(caller+": $user has unexpected type", "type", fmt.Sprintf("%T", ctx.Value("$user")))
		panic(caller + ": $user has wrong type in context")
	}
	return user
}

// RoleFromContext returns $role string from ctx. caller labels slog/panic.
func RoleFromContext(ctx context.Context, caller string) string {
	roleName, ok := ctx.Value("$role").(string)
	if !ok {
		slog.Error(caller+": $role has unexpected type", "type", fmt.Sprintf("%T", ctx.Value("$role")))
		panic(caller + ": $role has wrong type in context")
	}
	return roleName
}

// UserAndRoleFromContext returns $user and $role from ctx (same keys as auth layer).
func UserAndRoleFromContext(ctx context.Context, caller string) (User, string) {
	return UserFromContext(ctx, caller), RoleFromContext(ctx, caller)
}

// UserFromContextOptional returns (u, true) when $user is typed User; otherwise (zero, false).
func UserFromContextOptional(ctx context.Context) (User, bool) {
	u, ok := ctx.Value("$user").(User)
	return u, ok
}

// RoleFromContextOptional returns (role, true) when $role is a string; otherwise ("", false).
func RoleFromContextOptional(ctx context.Context) (string, bool) {
	s, ok := ctx.Value("$role").(string)
	return s, ok
}

// UserPresentInContext reports whether $user is present and typed as User.
func UserPresentInContext(ctx context.Context) bool {
	_, ok := UserFromContextOptional(ctx)
	return ok
}
