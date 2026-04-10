package p_otp

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

func init() {
	registerOtpAuthPages()
	registerOtpPreferencesPages()
}

func init() {
	lago.OnDBInit("p_otp.pages_bootstrap", func(d *gorm.DB) *gorm.DB {
		prefs := LoadPreferences(d)
		smsEnabled := prefs.SmsOtpTemplateId != "" || prefs.OtpTemplateId != ""
		emailEnabled := prefs.EmailOtpTemplateString != ""

		if !smsEnabled && !emailEnabled {
			return d
		}

		lago.RegistryPage.Patch("users.LoginPage", func(page components.PageInterface) components.PageInterface {
			if scaffold, ok := page.(*components.ShellAuthScaffold); ok {
				components.InsertChildAfter(scaffold,
					"users.AuthForm",
					func(*components.FormComponent[p_users.User]) *components.ButtonLink {
						return &components.ButtonLink{
							Label: "Forgot password?",
							Link:  lago.RoutePath("otp.ForgotPasswordRoute", nil),
						}
					})
				return scaffold
			}
			panic("Base page for login page was not ShellAuthScaffold")
		})
		return d
	})
}
