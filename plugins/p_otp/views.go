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

// PhoneOtpRequestHandler handles SMS OTP Generation.
func PhoneOtpRequestHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If authenticated, redirect
		if r.Context().Value("user") != nil {
			lago.NewRedirectView("users.ListRoute").ServeHTTP(w, r)
			return
		}

		if r.Method == http.MethodGet {
			v.RenderPage(w, r)
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			slog.Error("PhoneOtpRequestHandler: failed to parse form",
				"method", r.Method,
				"path", r.URL.Path,
				"error", err,
			)
			return
		}

		identifier, _ := values["Identifier"].(string)
		identifier = strings.TrimSpace(identifier)

		if identifier == "" {
			fieldErrors["Identifier"] = fmt.Errorf("Phone number is required.")
		}

		dbValue := r.Context().Value("$db")
		db, ok := dbValue.(*gorm.DB)
		if !ok || db == nil {
			slog.Error("PhoneOtpRequestHandler: missing or invalid *gorm.DB in context",
				"method", r.Method,
				"path", r.URL.Path,
				"identifier", identifier,
			)
			fieldErrors["Identifier"] = fmt.Errorf("Internal error. Please try again later.")
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		if !v.HasErrors(fieldErrors) {
			var count int64
			if err := db.Model(&p_users.User{}).Where("phone = ?", identifier).Count(&count).Error; err != nil {
				slog.Error("PhoneOtpRequestHandler: failed to count users by phone",
					"identifier", identifier,
					"error", err,
				)
				fieldErrors["Identifier"] = fmt.Errorf("Internal error. Please try again later.")
			} else if count == 0 {
				fieldErrors["Identifier"] = fmt.Errorf("No user found with this phone number.")
			} else {
				sent := SendSmsOtp(db, identifier)
				if sent {
					verifyPath, _ := getters.IfOr(lago.RoutePath("otp.OtpVerifyRoute", nil), r.Context(), "")
					successUrl := verifyPath + "?identifier=" + url.QueryEscape(identifier)
					lago.Redirect(w, r, successUrl)
					return
				} else {
					slog.Error("PhoneOtpRequestHandler: failed to send SMS OTP",
						"identifier", identifier,
					)
					fieldErrors["Identifier"] = fmt.Errorf("Failed to send OTP. Please check configuration.")
				}
			}
		}

		v.RenderWithErrors(w, r, fieldErrors, values)
	})
}

// EmailOtpRequestHandler handles Email OTP Generation.
func EmailOtpRequestHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value("user") != nil {
			lago.NewRedirectView("users.ListRoute").ServeHTTP(w, r)
			return
		}

		if r.Method == http.MethodGet {
			v.RenderPage(w, r)
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			slog.Error("EmailOtpRequestHandler: failed to parse form",
				"method", r.Method,
				"path", r.URL.Path,
				"error", err,
			)
			return
		}

		identifier, _ := values["Identifier"].(string)
		identifier = strings.TrimSpace(identifier)

		if identifier == "" {
			fieldErrors["Identifier"] = fmt.Errorf("Email address is required.")
		}

		dbValue := r.Context().Value("$db")
		db, ok := dbValue.(*gorm.DB)
		if !ok || db == nil {
			slog.Error("EmailOtpRequestHandler: missing or invalid *gorm.DB in context",
				"method", r.Method,
				"path", r.URL.Path,
				"identifier", identifier,
			)
			fieldErrors["Identifier"] = fmt.Errorf("Internal error. Please try again later.")
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		if !v.HasErrors(fieldErrors) {
			var count int64
			if err := db.Model(&p_users.User{}).Where("email = ?", identifier).Count(&count).Error; err != nil {
				slog.Error("EmailOtpRequestHandler: failed to count users by email",
					"identifier", identifier,
					"error", err,
				)
				fieldErrors["Identifier"] = fmt.Errorf("Internal error. Please try again later.")
			} else if count == 0 {
				fieldErrors["Identifier"] = fmt.Errorf("No user found with this email.")
			} else {
				sent := SendEmailOtp(db, identifier)
				if sent {
					verifyPath, _ := getters.IfOr(lago.RoutePath("otp.OtpVerifyRoute", nil), r.Context(), "")
					successUrl := verifyPath + "?identifier=" + url.QueryEscape(identifier)
					lago.Redirect(w, r, successUrl)
					return
				} else {
					slog.Error("EmailOtpRequestHandler: failed to send email OTP",
						"identifier", identifier,
					)
					fieldErrors["Identifier"] = fmt.Errorf("Failed to send OTP. Please check configuration.")
				}
			}
		}

		v.RenderWithErrors(w, r, fieldErrors, values)
	})
}

// OtpVerifyHandler verifies the code and logs the user in.
func OtpVerifyHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identifier := r.URL.Query().Get("identifier")
		if identifier == "" {
			slog.Warn("OtpVerifyHandler: missing identifier in query",
				"method", r.Method,
				"path", r.URL.Path,
			)
			lago.NewRedirectView("users.LoginRoute").ServeHTTP(w, r)
			return
		}

		if r.Method == http.MethodGet {
			ctx := context.WithValue(r.Context(), "$in", map[string]any{
				"Identifier": identifier,
			})
			r = r.WithContext(ctx)
			v.RenderPage(w, r)
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			slog.Error("OtpVerifyHandler: failed to parse form",
				"method", r.Method,
				"path", r.URL.Path,
				"identifier", identifier,
				"error", err,
			)
			return
		}

		newPassword, _ := values["NewPassword"].(string)
		newPassword2, _ := values["NewPassword2"].(string)
		newPassword = strings.TrimSpace(newPassword)
		newPassword2 = strings.TrimSpace(newPassword2)

		if newPassword == "" {
			fieldErrors["NewPassword"] = fmt.Errorf("New password is required.")
		}
		if newPassword2 == "" {
			fieldErrors["NewPassword2"] = fmt.Errorf("Please confirm your new password.")
		}
		if newPassword != "" && newPassword2 != "" && newPassword != newPassword2 {
			fieldErrors["NewPassword2"] = fmt.Errorf("Passwords do not match.")
		}

		otp, _ := values["Otp"].(string)
		otp = strings.TrimSpace(otp)

		if otp == "" {
			fieldErrors["Otp"] = fmt.Errorf("OTP is required.")
		} else if len(otp) != 6 {
			fieldErrors["Otp"] = fmt.Errorf("OTP must be 6 digits.")
		}

		dbValue := r.Context().Value("$db")
		db, ok := dbValue.(*gorm.DB)
		if !ok || db == nil {
			slog.Error("OtpVerifyHandler: missing or invalid *gorm.DB in context",
				"method", r.Method,
				"path", r.URL.Path,
				"identifier", identifier,
			)
			fieldErrors["Otp"] = fmt.Errorf("Internal error. Please try again later.")
			values["Identifier"] = identifier
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		if v.HasErrors(fieldErrors) {
			values["Identifier"] = identifier
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		if !VerifyOTP(identifier, otp) {
			fieldErrors["Otp"] = fmt.Errorf("Invalid OTP.")
			values["Identifier"] = identifier
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		var user p_users.User
		err = db.Where("phone = ? OR email = ?", identifier, identifier).First(&user).Error
		if err == nil {
			user.Password = []byte(newPassword)
			if err := db.Save(&user).Error; err != nil {
				slog.Error("OtpVerifyHandler: failed to update password",
					"identifier", identifier,
					"error", err,
				)
				fieldErrors["NewPassword"] = fmt.Errorf("Could not update password. Please try again.")
				values["Identifier"] = identifier
				v.RenderWithErrors(w, r, fieldErrors, values)
				return
			}
			user.Login(w)
			lago.NewRedirectView("users.LoginSuccessRoute").ServeHTTP(w, r)
			return
		}
		if err == gorm.ErrRecordNotFound {
			slog.Warn("OtpVerifyHandler: user not found for identifier",
				"identifier", identifier,
			)
			fieldErrors["Otp"] = fmt.Errorf("User not found.")
		} else {
			slog.Error("OtpVerifyHandler: database error while loading user",
				"identifier", identifier,
				"error", err,
			)
			fieldErrors["Otp"] = fmt.Errorf("Internal error. Please try again later.")
		}

		// Keep identifier around so form URL resolves correctly on re-render
		values["Identifier"] = identifier
		v.RenderWithErrors(w, r, fieldErrors, values)
	})
}

func init() {
	// SMS OTP Request
	lago.RegistryView.Register("otp.PhoneOtpRequestView",
		lago.GetPageView("otp.PhoneOtpRequestForm").
			WithMethod(http.MethodPost, PhoneOtpRequestHandler))

	// Email OTP Request
	lago.RegistryView.Register("otp.EmailOtpRequestView",
		lago.GetPageView("otp.EmailOtpRequestForm").
			WithMethod(http.MethodPost, EmailOtpRequestHandler))

	// OTP Verify
	lago.RegistryView.Register("otp.OtpVerifyView",
		lago.GetPageView("otp.OtpVerifyForm").
			WithMethod(http.MethodGet, OtpVerifyHandler).
			WithMethod(http.MethodPost, OtpVerifyHandler))

	// OTP Preferences
	lago.RegistryView.Register("otp.OTPPreferencesView",
		views.SingletonView[OTPPreferences](lago.RoutePath("otp.OTPPreferencesRoute", nil))(
			lago.GetPageView("otp.OTPPreferencesForm")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("users.role", p_users.RoleAuthorizationMiddleware([]string{"superuser"})))
}
