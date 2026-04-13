package p_otp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func redirectToRoute(w http.ResponseWriter, r *http.Request, routeKey string, args ...map[string]getters.Getter[any]) bool {
	var routeArgs map[string]getters.Getter[any]
	if len(args) > 0 {
		routeArgs = args[0]
	}
	urlValue, err := getters.IfOr(lago.RoutePath(routeKey, routeArgs), r.Context(), "")
	if err != nil || urlValue == "" {
		http.NotFound(w, r)
		return false
	}
	views.HtmxRedirect(w, r, urlValue, http.StatusMovedPermanently)
	return true
}

func phoneOtpRequestHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p_users.UserPresentInContext(r.Context()) {
			_ = redirectToRoute(w, r, "users.ListRoute")
			return
		}

		if r.Method == http.MethodGet {
			v.RenderPage(w, r)
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": err})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		identifier, _ := values["Identifier"].(string)
		identifier = strings.TrimSpace(identifier)
		if identifier == "" {
			fieldErrors["Identifier"] = fmt.Errorf("phone number is required")
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("PhoneOtpRequestHandler: db from context", "error", dberr)
			fieldErrors["Identifier"] = fmt.Errorf("internal error. please try again later")
		}

		if len(fieldErrors) == 0 {
			count, err := gorm.G[p_users.User](db).Where("phone = ?", identifier).Count(r.Context(), "*")
			if err != nil {
				slog.Error("PhoneOtpRequestHandler: failed to count users by phone", "identifier", identifier, "error", err)
				fieldErrors["Identifier"] = fmt.Errorf("internal error. please try again later")
			} else if count == 0 {
				fieldErrors["Identifier"] = fmt.Errorf("no user found with this phone number")
			} else if SendSmsOtp(db, identifier) {
				verifyPath, _ := getters.IfOr(lago.RoutePath("otp.OtpVerifyRoute", nil), r.Context(), "")
				views.HtmxRedirect(w, r, verifyPath+"?identifier="+url.QueryEscape(identifier), http.StatusMovedPermanently)
				return
			} else {
				slog.Error("PhoneOtpRequestHandler: failed to send SMS OTP", "identifier", identifier)
				fieldErrors["Identifier"] = fmt.Errorf("failed to send OTP. please check configuration")
			}
		}

		ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func emailOtpRequestHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p_users.UserPresentInContext(r.Context()) {
			_ = redirectToRoute(w, r, "users.ListRoute")
			return
		}

		if r.Method == http.MethodGet {
			v.RenderPage(w, r)
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": err})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		identifier, _ := values["Identifier"].(string)
		identifier = strings.TrimSpace(identifier)
		if identifier == "" {
			fieldErrors["Identifier"] = fmt.Errorf("email address is required")
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("EmailOtpRequestHandler: db from context", "error", dberr)
			fieldErrors["Identifier"] = fmt.Errorf("internal error. please try again later")
		}

		if len(fieldErrors) == 0 {
			count, err := gorm.G[p_users.User](db).Where("email = ?", identifier).Count(r.Context(), "*")
			if err != nil {
				slog.Error("EmailOtpRequestHandler: failed to count users by email", "identifier", identifier, "error", err)
				fieldErrors["Identifier"] = fmt.Errorf("internal error. please try again later")
			} else if count == 0 {
				fieldErrors["Identifier"] = fmt.Errorf("no user found with this email")
			} else if SendEmailOtp(db, identifier) {
				verifyPath, _ := getters.IfOr(lago.RoutePath("otp.OtpVerifyRoute", nil), r.Context(), "")
				views.HtmxRedirect(w, r, verifyPath+"?identifier="+url.QueryEscape(identifier), http.StatusMovedPermanently)
				return
			} else {
				slog.Error("EmailOtpRequestHandler: failed to send email OTP", "identifier", identifier)
				fieldErrors["Identifier"] = fmt.Errorf("failed to send OTP. please check configuration")
			}
		}

		ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func otpVerifyHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identifier := r.URL.Query().Get("identifier")
		if identifier == "" {
			_ = redirectToRoute(w, r, "users.LoginRoute")
			return
		}

		if r.Method == http.MethodGet {
			ctx := context.WithValue(r.Context(), "$in", map[string]any{
				"Identifier": identifier,
			})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			values["Identifier"] = identifier
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": err})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		newPassword, _ := values["NewPassword"].(string)
		newPassword2, _ := values["NewPassword2"].(string)
		newPassword = strings.TrimSpace(newPassword)
		newPassword2 = strings.TrimSpace(newPassword2)
		if newPassword == "" {
			fieldErrors["NewPassword"] = fmt.Errorf("new password is required")
		}
		if newPassword2 == "" {
			fieldErrors["NewPassword2"] = fmt.Errorf("please confirm your new password")
		}
		if newPassword != "" && newPassword2 != "" && newPassword != newPassword2 {
			fieldErrors["NewPassword2"] = fmt.Errorf("passwords do not match")
		}

		otp, _ := values["Otp"].(string)
		otp = strings.TrimSpace(otp)
		if otp == "" {
			fieldErrors["Otp"] = fmt.Errorf("OTP is required")
		} else if len(otp) != 6 {
			fieldErrors["Otp"] = fmt.Errorf("OTP must be 6 digits")
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("OtpVerifyHandler: db from context", "error", dberr)
			fieldErrors["Otp"] = fmt.Errorf("internal error. please try again later")
		}

		if len(fieldErrors) != 0 {
			values["Identifier"] = identifier
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		if !VerifyOTP(identifier, otp) {
			fieldErrors["Otp"] = fmt.Errorf("invalid OTP")
			values["Identifier"] = identifier
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		user, err := gorm.G[p_users.User](db).Where("phone = ? OR email = ?", identifier, identifier).First(r.Context())
		if err == nil {
			user.Password = []byte(newPassword)
			if err := db.Save(&user).Error; err != nil {
				slog.Error("OtpVerifyHandler: failed to update password", "identifier", identifier, "error", err)
				fieldErrors["NewPassword"] = fmt.Errorf("could not update password. please try again")
				values["Identifier"] = identifier
				ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
				v.RenderPage(w, r.WithContext(ctx))
				return
			}
			user.Login(w, r)
			_ = redirectToRoute(w, r, "users.LoginRoute")
			return
		}

		if err == gorm.ErrRecordNotFound {
			slog.Warn("OtpVerifyHandler: user not found for identifier", "identifier", identifier)
			fieldErrors["Otp"] = fmt.Errorf("user not found")
		} else {
			slog.Error("OtpVerifyHandler: database error while loading user", "identifier", identifier, "error", err)
			fieldErrors["Otp"] = fmt.Errorf("internal error. please try again later")
		}

		values["Identifier"] = identifier
		ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func init() {
	lago.RegistryView.Register("otp.ForgotPasswordView",
		lago.GetPageView("otp.ForgotPasswordPage"))

	lago.RegistryView.Register("otp.PhoneOtpRequestView",
		lago.GetPageView("otp.PhoneOtpRequestForm").
			WithLayer("users.optional_auth", p_users.OptionalAuthLayer{}).
			WithLayer("otp.phone_get", views.MethodLayer{
				Method:  http.MethodGet,
				Handler: phoneOtpRequestHandler,
			}).
			WithLayer("otp.phone_post", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: phoneOtpRequestHandler,
			}))

	lago.RegistryView.Register("otp.EmailOtpRequestView",
		lago.GetPageView("otp.EmailOtpRequestForm").
			WithLayer("users.optional_auth", p_users.OptionalAuthLayer{}).
			WithLayer("otp.email_get", views.MethodLayer{
				Method:  http.MethodGet,
				Handler: emailOtpRequestHandler,
			}).
			WithLayer("otp.email_post", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: emailOtpRequestHandler,
			}))

	lago.RegistryView.Register("otp.OtpVerifyView",
		lago.GetPageView("otp.OtpVerifyForm").
			WithLayer("users.optional_auth", p_users.OptionalAuthLayer{}).
			WithLayer("otp.verify_get", views.MethodLayer{
				Method:  http.MethodGet,
				Handler: otpVerifyHandler,
			}).
			WithLayer("otp.verify_post", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: otpVerifyHandler,
			}))

	lago.RegistryView.Register("otp.OTPPreferencesView",
		lago.GetPageView("otp.OTPPreferencesForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("users.role", p_users.RoleAuthorizationLayer{Roles: []string{"superuser"}}).
			WithLayer("otp.preferences", views.LayerSingleton[OTPPreferences]{
				SuccessURL: lago.RoutePath("otp.OTPPreferencesRoute", nil),
			}))
}
