package p_totschool_otp

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/views"

	_ "github.com/lariv-in/p_otp"
	"github.com/lariv-in/p_users"
)

// otpViewsWithRoleMiddleware are the users plugin views that use "users.role" middleware.
var otpViewsWithRoleMiddleware = []string{
	"otp.OTPPreferencesView",
}

// Patch "users.role" middleware to allow totschool_admin in addition to existing roles.
func patcher(current views.Middleware) views.Middleware {
	return p_users.RoleAuthorizationMiddleware([]string{"", "totschool_admin"})
}

func init() {
	for _, key := range otpViewsWithRoleMiddleware {
		if v, ok := lago.RegistryView.Get(key); ok {
			v.PatchMiddleware("users.role", patcher)
		}
	}
}
