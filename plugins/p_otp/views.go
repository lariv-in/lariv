package p_otp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/lariv-in/components"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

// ensureBaseForm prepares the page/form components parsing and returns true if there are parsing errors.
func ensureBaseForm(w http.ResponseWriter, r *http.Request, page components.PageInterface) (map[string]any, map[string]error, bool) {
	forms := components.FindChildren[components.FormComponent](page.(components.ParentInterface))
	if len(forms) == 0 {
		http.Error(w, "Internal Server Error: No form found", http.StatusInternalServerError)
		return nil, nil, true
	}
	form := forms[0]
	values, fieldErrors, err := form.ParseForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, nil, true
	}
	return values, fieldErrors, false
}

func renderWithErrors(w http.ResponseWriter, r *http.Request, page components.PageInterface, fieldErrors map[string]error, values map[string]any) {
	ctx := r.Context()
	for name, fieldErr := range fieldErrors {
		if fieldErr != nil {
			ctx = context.WithValue(ctx, "$error."+name, fieldErr)
		}
	}
	for name, value := range values {
		ctx = context.WithValue(ctx, "$in."+name, value)
	}
	page.Build(ctx).Render(w)
}

func hasErrors(errs map[string]error) bool {
	for _, err := range errs {
		if err != nil {
			return true
		}
	}
	return false
}

// PhoneOtpRequestHandler handles SMS OTP Generation.
func PhoneOtpRequestHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := v.GetPage()

		// If authenticated, redirect
		if r.Context().Value("user") != nil {
			lago.NewRedirectView("users.ListRoute").ServeHTTP(w, r)
			return
		}

		if r.Method == http.MethodGet {
			page.Build(r.Context()).Render(w)
			return
		}

		values, fieldErrors, failed := ensureBaseForm(w, r, page)
		if failed {
			return
		}

		identifier, _ := values["identifier"].(string)
		identifier = strings.TrimSpace(identifier)

		if identifier == "" {
			fieldErrors["identifier"] = fmt.Errorf("Phone number is required.")
		}

		db := r.Context().Value("$db").(*gorm.DB)

		if !hasErrors(fieldErrors) {
			var count int64
			db.Model(&p_users.User{}).Where("phone = ?", identifier).Count(&count)
			if count == 0 {
				fieldErrors["identifier"] = fmt.Errorf("No user found with this phone number.")
			} else {
				sent := SendSmsOtp(db, identifier)
				if sent {
					// Redirect to verify with identifier safely passed as param
					successUrl := fmt.Sprintf("/otp/verify/?identifier=%v", identifier)
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

		renderWithErrors(w, r, page, fieldErrors, values)
	})
}

// EmailOtpRequestHandler handles Email OTP Generation.
func EmailOtpRequestHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := v.GetPage()

		if r.Context().Value("user") != nil {
			lago.NewRedirectView("users.ListRoute").ServeHTTP(w, r)
			return
		}

		if r.Method == http.MethodGet {
			page.Build(r.Context()).Render(w)
			return
		}

		values, fieldErrors, failed := ensureBaseForm(w, r, page)
		if failed {
			return
		}

		identifier, _ := values["identifier"].(string)
		identifier = strings.TrimSpace(identifier)

		if identifier == "" {
			fieldErrors["identifier"] = fmt.Errorf("Email address is required.")
		}

		db := r.Context().Value("$db").(*gorm.DB)

		if !hasErrors(fieldErrors) {
			var count int64
			// simplistic check, since model has Email uniqueIndex
			db.Model(&p_users.User{}).Where("email = ?", identifier).Count(&count)
			if count == 0 {
				fieldErrors["identifier"] = fmt.Errorf("No user found with this email.")
			} else {
				sent := SendEmailOtp(db, identifier)
				if sent {
					successUrl := fmt.Sprintf("/otp/verify/?identifier=%v", identifier)
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

		renderWithErrors(w, r, page, fieldErrors, values)
	})
}

// OtpVerifyHandler verifies the code and logs the user in.
func OtpVerifyHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := v.GetPage()

		identifier := r.URL.Query().Get("identifier")
		if identifier == "" {
			lago.NewRedirectView("users.LoginRoute").ServeHTTP(w, r)
			return
		}

		if r.Method == http.MethodGet {
			ctx := context.WithValue(r.Context(), "$in.identifier", identifier)
			page.Build(ctx).Render(w)
			return
		}

		values, fieldErrors, failed := ensureBaseForm(w, r, page)
		if failed {
			return
		}

		otp, _ := values["otp"].(string)
		otp = strings.TrimSpace(otp)

		if otp == "" {
			fieldErrors["otp"] = fmt.Errorf("OTP is required.")
		} else if len(otp) != 6 {
			fieldErrors["otp"] = fmt.Errorf("OTP must be 6 digits.")
		} else if !VerifyOTP(identifier, otp) {
			fieldErrors["otp"] = fmt.Errorf("Invalid OTP.")
		}

		db := r.Context().Value("$db").(*gorm.DB)

		if !hasErrors(fieldErrors) {
			var user p_users.User
			// Try finding by phone or email
			err := db.Where("phone = ? OR email = ?", identifier, identifier).First(&user).Error
			if err == nil {
				user.Login(w)
				successUrl := "/users/" // Assuming users:list mapped to /users/
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

		// Keep identifier around in form context
		values["identifier"] = identifier
		renderWithErrors(w, r, page, fieldErrors, values)
	})
}

// OTPPreferencesHandler modifies OTP preferences.
func OTPPreferencesHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := v.GetPage()
		db := r.Context().Value("$db").(*gorm.DB)

		prefs := LoadPreferences(db)

		if r.Method == http.MethodGet {
			// Populate the form with current preferences
			ctx := r.Context()
			ctx = context.WithValue(ctx, "$in.msg91_auth_key", prefs.Msg91AuthKey)
			ctx = context.WithValue(ctx, "$in.sms_otp_template_id", prefs.SmsOtpTemplateId)
			ctx = context.WithValue(ctx, "$in.otp_template_id", prefs.OtpTemplateId)
			ctx = context.WithValue(ctx, "$in.sms_otp_field_name", prefs.SmsOtpFieldName)
			ctx = context.WithValue(ctx, "$in.sms_otp_extra_fields", prefs.SmsOtpExtraFields)
			ctx = context.WithValue(ctx, "$in.email_otp_template_string", prefs.EmailOtpTemplateString)
			page.Build(ctx).Render(w)
			return
		}

		values, fieldErrors, failed := ensureBaseForm(w, r, page)
		if failed {
			return
		}

		if val, ok := values["msg91_auth_key"].(string); ok {
			prefs.Msg91AuthKey = strings.TrimSpace(val)
		}
		if val, ok := values["sms_otp_template_id"].(string); ok {
			prefs.SmsOtpTemplateId = strings.TrimSpace(val)
		}
		if val, ok := values["otp_template_id"].(string); ok {
			prefs.OtpTemplateId = strings.TrimSpace(val)
		}

		fieldName, _ := values["sms_otp_field_name"].(string)
		fieldName = strings.TrimSpace(fieldName)
		if fieldName == "" {
			fieldName = "otp"
		}
		prefs.SmsOtpFieldName = fieldName

		extraFieldsStr, _ := values["sms_otp_extra_fields"].(string)
		extraFieldsStr = strings.TrimSpace(extraFieldsStr)
		if extraFieldsStr != "" {
			var dummy map[string]any
			if err := json.Unmarshal([]byte(extraFieldsStr), &dummy); err != nil {
				fieldErrors["sms_otp_extra_fields"] = fmt.Errorf("Invalid JSON format")
			}
		}
		prefs.SmsOtpExtraFields = extraFieldsStr

		if val, ok := values["email_otp_template_string"].(string); ok {
			prefs.EmailOtpTemplateString = strings.TrimSpace(val)
		}

		if !hasErrors(fieldErrors) {
			db.Save(&prefs)

			successUrl := "/otp/preferences/"
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", successUrl)
				w.WriteHeader(http.StatusOK)
			} else {
				http.Redirect(w, r, successUrl, http.StatusSeeOther)
			}
			return
		}

		renderWithErrors(w, r, page, fieldErrors, values)
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
			WithMethod(http.MethodPost, OtpVerifyHandler))

	// OTP Preferences
	lago.RegistryView.Register("otp.OTPPreferencesView",
		p_users.AuthMiddleware(
			lago.GetPageView("otp.OTPPreferencesForm").
				WithMethod(http.MethodPost, OTPPreferencesHandler)))
}
