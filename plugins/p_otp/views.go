package p_otp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

// PhoneOtpRequestHandler handles SMS OTP Generation.
func PhoneOtpRequestHandler(v views.View) http.Handler {
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
			return
		}

		identifier, _ := values["Identifier"].(string)
		identifier = strings.TrimSpace(identifier)

		if identifier == "" {
			fieldErrors["identifier"] = fmt.Errorf("Phone number is required.")
		}

		db := r.Context().Value("$db").(*gorm.DB)

		if !views.HasErrors(fieldErrors) {
			var count int64
			db.Model(&p_users.User{}).Where("phone = ?", identifier).Count(&count)
			if count == 0 {
				fieldErrors["identifier"] = fmt.Errorf("No user found with this phone number.")
			} else {
				sent := SendSmsOtp(db, identifier)
				if sent {
					successUrl := "/otp/verify/?identifier=" + url.QueryEscape(identifier)
					if r.Header.Get("HX-Request") == "true" {
						w.Header().Set("HX-Redirect", successUrl)
						w.WriteHeader(http.StatusOK)
					} else {
						http.Redirect(w, r, successUrl, http.StatusSeeOther)
					}
					return
				} else {
					fieldErrors["identifier"] = fmt.Errorf("Failed to send OTP. Please check configuration.")
				}
			}
		}

		v.RenderWithErrors(w, r, fieldErrors, values)
	})
}

// EmailOtpRequestHandler handles Email OTP Generation.
func EmailOtpRequestHandler(v views.View) http.Handler {
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
			return
		}

		identifier, _ := values["Identifier"].(string)
		identifier = strings.TrimSpace(identifier)

		if identifier == "" {
			fieldErrors["identifier"] = fmt.Errorf("Email address is required.")
		}

		db := r.Context().Value("$db").(*gorm.DB)

		if !views.HasErrors(fieldErrors) {
			var count int64
			db.Model(&p_users.User{}).Where("email = ?", identifier).Count(&count)
			if count == 0 {
				fieldErrors["identifier"] = fmt.Errorf("No user found with this email.")
			} else {
				sent := SendEmailOtp(db, identifier)
				if sent {
					successUrl := "/otp/verify/?identifier=" + url.QueryEscape(identifier)
					if r.Header.Get("HX-Request") == "true" {
						w.Header().Set("HX-Redirect", successUrl)
						w.WriteHeader(http.StatusOK)
					} else {
						http.Redirect(w, r, successUrl, http.StatusSeeOther)
					}
					return
				} else {
					fieldErrors["identifier"] = fmt.Errorf("Failed to send OTP. Please check configuration.")
				}
			}
		}

		v.RenderWithErrors(w, r, fieldErrors, values)
	})
}

// OtpVerifyHandler verifies the code and logs the user in.
func OtpVerifyHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identifier := r.URL.Query().Get("identifier")
		if identifier == "" {
			lago.NewRedirectView("users.LoginRoute").ServeHTTP(w, r)
			return
		}

		if r.Method == http.MethodGet {
			ctx := context.WithValue(r.Context(), "$in", map[string]any{
				"identifier": identifier,
			})
			r = r.WithContext(ctx)
			v.RenderPage(w, r)
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			return
		}

		otp, _ := values["Otp"].(string)
		otp = strings.TrimSpace(otp)

		if otp == "" {
			fieldErrors["otp"] = fmt.Errorf("OTP is required.")
		} else if len(otp) != 6 {
			fieldErrors["otp"] = fmt.Errorf("OTP must be 6 digits.")
		} else if !VerifyOTP(identifier, otp) {
			fieldErrors["otp"] = fmt.Errorf("Invalid OTP.")
		}

		db := r.Context().Value("$db").(*gorm.DB)

		if !views.HasErrors(fieldErrors) {
			var user p_users.User
			err := db.Where("phone = ? OR email = ?", identifier, identifier).First(&user).Error
			if err == nil {
				user.Login(w)
				successUrl := "/users/"
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Redirect", successUrl)
					w.WriteHeader(http.StatusOK)
				} else {
					http.Redirect(w, r, successUrl, http.StatusSeeOther)
				}
				return
			} else {
				fieldErrors["otp"] = fmt.Errorf("User not found.")
			}
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
		p_users.AuthMiddleware(
			p_users.RoleAuthorizationMiddleware([]string{"superuser"})(
				views.SingletonView(OTPPreferences{}, lago.RoutePathGetter("otp.OTPPreferencesRoute"))(
					lago.GetPageView("otp.OTPPreferencesForm")))))
}
