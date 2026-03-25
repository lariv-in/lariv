package p_otp

import (
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryPage.Register("otp.PhoneOtpRequestForm", components.ShellAuthScaffold{
		Children: []components.PageInterface{
			components.ContainerColumn{
				Children: []components.PageInterface{
					components.FieldTitle{Getter: getters.GetterStatic("Login via SMS")},
					components.FormComponent[map[string]string]{
						Url:    lago.GetterRoutePath("otp.PhoneOtpRequestRoute", nil),
						Method: http.MethodPost,
						ChildrenInput: []components.PageInterface{
							components.ContainerError{
								Error: getters.GetterKey[error]("$error.Identifier"),
								Children: []components.PageInterface{
									components.InputPhone{
										Name:     "Identifier",
										Label:    "Phone Number",
										Required: true,
										Getter:   getters.GetterKey[string]("$in.Identifier"),
									},
								},
							},
						},
						ChildrenAction: []components.PageInterface{
							components.ButtonSubmit{
								Label:   "Send OTP",
								Classes: "w-full",
							},
						},
					},
					components.ContainerRow{
						Classes: "text-center mt-4",
						Children: []components.PageInterface{
							components.ButtonLink{
								Label: "Back to Login",
								Link:  lago.GetterRoutePath("users.LoginRoute", nil),
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("otp.EmailOtpRequestForm", components.ShellAuthScaffold{
		Children: []components.PageInterface{
			components.ContainerColumn{
				Classes: "w-80",
				Children: []components.PageInterface{
					components.FieldTitle{Getter: getters.GetterStatic("Login via Email")},
					components.FormComponent[map[string]string]{
						Url:    lago.GetterRoutePath("otp.EmailOtpRequestRoute", nil),
						Method: http.MethodPost,
						ChildrenInput: []components.PageInterface{
							components.ContainerError{
								Error: getters.GetterKey[error]("$error.Identifier"),
								Children: []components.PageInterface{
									components.InputEmail{
										Name:     "Identifier",
										Label:    "Email Address",
										Required: true,
										Getter:   getters.GetterKey[string]("$in.Identifier"),
									},
								},
							},
						},
						ChildrenAction: []components.PageInterface{
							components.ButtonSubmit{
								Label:   "Send OTP",
								Classes: "w-full",
							},
						},
					},
					components.ContainerRow{
						Classes: "text-center mt-4",
						Children: []components.PageInterface{
							components.ButtonLink{
								Label: "Back to Login",
								Link:  lago.GetterRoutePath("users.LoginRoute", nil),
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("otp.OtpVerifyForm", components.ShellAuthScaffold{
		Children: []components.PageInterface{
			components.ContainerColumn{
				Classes: "w-80",
				Children: []components.PageInterface{
					components.FieldTitle{Getter: getters.GetterStatic("Verify OTP")},
					components.FieldText{
						Classes: "text-sm text-gray-600 mb-2",
						Getter:  getters.GetterStatic("Enter the code we sent and choose a new password."),
					},
					components.FormComponent[map[string]string]{
						Url:    getters.GetterFormat("%v?identifier=%v", getters.GetterAny(lago.GetterRoutePath("otp.OtpVerifyRoute", nil)), getters.GetterAny(getters.GetterQueryEscape(getters.GetterKey[string]("$in.Identifier")))),
						Method: http.MethodPost,
						ChildrenInput: []components.PageInterface{
							components.ContainerError{
								Error: getters.GetterKey[error]("$error.Otp"),
								Children: []components.PageInterface{
									components.InputText{
										Name:     "Otp",
										Label:    "OTP",
										Required: true,
										Getter:   getters.GetterKey[string]("$in.Otp"),
									},
								},
							},
							components.ContainerError{
								Error: getters.GetterKey[error]("$error.NewPassword"),
								Children: []components.PageInterface{
									components.InputPassword{
										Name:     "NewPassword",
										Label:    "New password",
										Required: true,
										Getter:   getters.GetterKey[string]("$in.NewPassword"),
									},
								},
							},
							components.ContainerError{
								Error: getters.GetterKey[error]("$error.NewPassword2"),
								Children: []components.PageInterface{
									components.InputPassword{
										Name:     "NewPassword2",
										Label:    "Confirm new password",
										Required: true,
										Getter:   getters.GetterKey[string]("$in.NewPassword2"),
									},
								},
							},
						},
						ChildrenAction: []components.PageInterface{
							components.ButtonSubmit{
								Label:   "Verify & Login",
								Classes: "w-full",
							},
						},
					},
					components.ContainerRow{
						Classes: "text-center mt-4",
						Children: []components.PageInterface{
							components.ButtonLink{
								Label: "Cancel",
								Link:  lago.GetterRoutePath("users.LoginRoute", nil),
							},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("otp.OTPPreferencesMenu", components.SidebarMenu{
		Title: getters.GetterStatic("OTP Preferences"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to Home"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Preferences"),
				Url:   lago.GetterRoutePath("otp.OTPPreferencesRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("otp.OTPPreferencesForm", components.ShellScaffold{
		Page: components.Page{Roles: []string{"superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "otp.OTPPreferencesMenu"},
		},
		Children: []components.PageInterface{
			components.FormComponent[OTPPreferences]{
				Url:      lago.GetterRoutePath("otp.OTPPreferencesRoute", nil),
				Title:    "OTP Preferences",
				Subtitle: "Configure OTP settings for SMS and Email",
				Method:   http.MethodPost,
				ChildrenInput: []components.PageInterface{
					components.FieldText{
						Classes: "text-lg font-semibold mt-4",
						Getter:  getters.GetterStatic("SMS OTP Settings"),
					},
					components.ContainerError{
						Error: getters.GetterKey[error]("$error.Msg91AuthKey"),
						Children: []components.PageInterface{
							components.InputText{
								Name:   "Msg91AuthKey",
								Label:  "MSG91 Auth Key",
								Getter: getters.GetterKey[string]("$in.Msg91AuthKey"),
							},
						},
					},
					components.ContainerRow{
						Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
						Children: []components.PageInterface{
							components.ContainerError{
								Error: getters.GetterKey[error]("$error.SmsOtpTemplateId"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmsOtpTemplateId",
										Label:  "SMS OTP Template ID",
										Getter: getters.GetterKey[string]("$in.SmsOtpTemplateId"),
									},
								},
							},
							components.ContainerError{
								Error: getters.GetterKey[error]("$error.OtpTemplateId"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "OtpTemplateId",
										Label:  "General OTP Template ID (Fallback)",
										Getter: getters.GetterKey[string]("$in.OtpTemplateId"),
									},
								},
							},
						},
					},
					components.ContainerError{
						Error: getters.GetterKey[error]("$error.SmsOtpFieldName"),
						Children: []components.PageInterface{
							components.InputText{
								Name:   "SmsOtpFieldName",
								Label:  "SMS OTP Field Name",
								Getter: getters.GetterKey[string]("$in.SmsOtpFieldName"),
							},
						},
					},
					components.ContainerError{
						Error: getters.GetterKey[error]("$error.SmsOtpExtraFields"),
						Children: []components.PageInterface{
							components.InputText{
								Name:   "SmsOtpExtraFields",
								Label:  "SMS OTP Extra Fields (JSON)",
								Getter: getters.GetterKey[string]("$in.SmsOtpExtraFields"),
							},
						},
					},
					components.FieldText{
						Classes: "text-lg font-semibold mt-4",
						Getter:  getters.GetterStatic("Email OTP Settings"),
					},
					components.ContainerError{
						Error: getters.GetterKey[error]("$error.EmailOtpTemplateString"),
						Children: []components.PageInterface{
							components.InputText{
								Name:   "EmailOtpTemplateString",
								Label:  "Email OTP Template String",
								Getter: getters.GetterKey[string]("$in.EmailOtpTemplateString"),
							},
						},
					},
					components.ContainerRow{
						Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
						Children: []components.PageInterface{
							components.ContainerError{
								Error: getters.GetterKey[error]("$error.SmtpHost"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmtpHost",
										Label:  "SMTP Host",
										Getter: getters.GetterKey[string]("$in.SmtpHost"),
									},
								},
							},
							components.ContainerError{
								Error: getters.GetterKey[error]("$error.SmtpPort"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmtpPort",
										Label:  "SMTP Port",
										Getter: getters.GetterKey[string]("$in.SmtpPort"),
									},
								},
							},
						},
					},
					components.ContainerRow{
						Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
						Children: []components.PageInterface{
							components.ContainerError{
								Error: getters.GetterKey[error]("$error.SmtpUsername"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmtpUsername",
										Label:  "SMTP Username",
										Getter: getters.GetterKey[string]("$in.SmtpUsername"),
									},
								},
							},
							components.ContainerError{
								Error: getters.GetterKey[error]("$error.SmtpPassword"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmtpPassword",
										Label:  "SMTP Password",
										Getter: getters.GetterKey[string]("$in.SmtpPassword"),
									},
								},
							},
						},
					},
					components.ContainerError{
						Error: getters.GetterKey[error]("$error.SmtpFrom"),
						Children: []components.PageInterface{
							components.InputText{
								Name:   "SmtpFrom",
								Label:  "SMTP From Address",
								Getter: getters.GetterKey[string]("$in.SmtpFrom"),
							},
						},
					},
				},
				ChildrenAction: []components.PageInterface{
					components.ButtonSubmit{
						Label: "Save Preferences",
					},
				},
			},
		},
	})
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
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
							Label: "Login with SMS OTP",
							Link:  lago.GetterRoutePath("otp.PhoneOtpRequestRoute", nil),
						}
					})
				components.InsertChildAfter(scaffold,
					"users.AuthForm",
					func(*components.FormComponent[p_users.User]) *components.ButtonLink {
						return &components.ButtonLink{
							Label: "Login with Email OTP",
							Link:  lago.GetterRoutePath("otp.EmailOtpRequestRoute", nil),
						}
					})
				return scaffold
			}
			panic("Base page for login page was not ShellAuthScaffold")
		})
		return d
	})
}
