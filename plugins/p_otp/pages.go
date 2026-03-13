package p_otp

import (
	"net/http"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryPage.Register("otp.PhoneOtpRequestForm", components.ShellAuthScaffold{
		Children: []components.PageInterface{
			components.ContainerColumn{
				Classes: "w-80",
				Children: []components.PageInterface{
					components.FieldTitle{Getter: getters.GetterStatic("Login via SMS")},
					components.FormComponent{
						Url:    getters.GetterStatic("/otp/login/sms/"),
						Method: http.MethodPost,
						ChildrenInput: []components.PageInterface{
							components.ContainerError{
								Error: getters.GetterKey("$error.Identifier"),
								Children: []components.PageInterface{
									components.InputPhone{
										Name:     "Identifier",
										Label:    "Phone Number",
										Required: true,
										Getter:   getters.GetterKey("$in.Identifier"),
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
								Link:  getters.GetterStatic("/users/login/"),
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
					components.FormComponent{
						Url:    getters.GetterStatic("/otp/login/email/"),
						Method: http.MethodPost,
						ChildrenInput: []components.PageInterface{
							components.ContainerError{
								Error: getters.GetterKey("$error.Identifier"),
								Children: []components.PageInterface{
									components.InputEmail{
										Name:     "Identifier",
										Label:    "Email Address",
										Required: true,
										Getter:   getters.GetterKey("$in.Identifier"),
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
								Link:  getters.GetterStatic("/users/login/"),
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
					components.FormComponent{
						Url:    getters.GetterFormat("/otp/verify/?identifier=%v", getters.GetterQueryEscape(getters.GetterKey("$in.Identifier"))),
						Method: http.MethodPost,
						ChildrenInput: []components.PageInterface{
							components.ContainerError{
								Error: getters.GetterKey("$error.Otp"),
								Children: []components.PageInterface{
									components.InputText{
										Name:     "Otp",
										Label:    "OTP",
										Required: true,
										Getter:   getters.GetterKey("$in.Otp"),
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
								Link:  getters.GetterStatic("/users/login/"),
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
			Url:   getters.GetterStatic("/apps/"),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Preferences"),
				Url:   getters.GetterStatic("/otp/preferences/"),
			},
		},
	})

	lago.RegistryPage.Register("otp.OTPPreferencesForm", components.ShellScaffold{
		Page: components.Page{RenderKeys: []string{"totschool_admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "otp.OTPPreferencesMenu"},
		},
		Children: []components.PageInterface{
			components.FormComponent{
				Url:      getters.GetterStatic("/otp/preferences/"),
				Title:    "OTP Preferences",
				Subtitle: "Configure OTP settings for SMS and Email",
				Method:   http.MethodPost,
				ChildrenInput: []components.PageInterface{
					components.FieldTitle{
						Getter: getters.GetterStatic("SMS OTP Settings"),
					},
					components.ContainerError{
						Error: getters.GetterKey("$error.Msg91AuthKey"),
						Children: []components.PageInterface{
							components.InputText{
								Name:   "Msg91AuthKey",
								Label:  "MSG91 Auth Key",
								Getter: getters.GetterKey("$in.Msg91AuthKey"),
							},
						},
					},
					components.ContainerRow{
						Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
						Children: []components.PageInterface{
							components.ContainerError{
								Error: getters.GetterKey("$error.SmsOtpTemplateId"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmsOtpTemplateId",
										Label:  "SMS OTP Template ID",
										Getter: getters.GetterKey("$in.SmsOtpTemplateId"),
									},
								},
							},
							components.ContainerError{
								Error: getters.GetterKey("$error.OtpTemplateId"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "OtpTemplateId",
										Label:  "General OTP Template ID (Fallback)",
										Getter: getters.GetterKey("$in.OtpTemplateId"),
									},
								},
							},
						},
					},
					components.ContainerError{
						Error: getters.GetterKey("$error.SmsOtpFieldName"),
						Children: []components.PageInterface{
							components.InputText{
								Name:   "SmsOtpFieldName",
								Label:  "SMS OTP Field Name",
								Getter: getters.GetterKey("$in.SmsOtpFieldName"),
							},
						},
					},
					components.ContainerError{
						Error: getters.GetterKey("$error.SmsOtpExtraFields"),
						Children: []components.PageInterface{
							components.InputText{
								Name:   "SmsOtpExtraFields",
								Label:  "SMS OTP Extra Fields (JSON)",
								Getter: getters.GetterKey("$in.SmsOtpExtraFields"),
							},
						},
					},
					components.FieldTitle{
						Getter: getters.GetterStatic("Email OTP Settings"),
					},
					components.ContainerError{
						Error: getters.GetterKey("$error.EmailOtpTemplateString"),
						Children: []components.PageInterface{
							components.InputText{
								Name:   "EmailOtpTemplateString",
								Label:  "Email OTP Template String",
								Getter: getters.GetterKey("$in.EmailOtpTemplateString"),
							},
						},
					},
					components.ContainerRow{
						Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
						Children: []components.PageInterface{
							components.ContainerError{
								Error: getters.GetterKey("$error.SmtpHost"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmtpHost",
										Label:  "SMTP Host",
										Getter: getters.GetterKey("$in.SmtpHost"),
									},
								},
							},
							components.ContainerError{
								Error: getters.GetterKey("$error.SmtpPort"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmtpPort",
										Label:  "SMTP Port",
										Getter: getters.GetterKey("$in.SmtpPort"),
									},
								},
							},
						},
					},
					components.ContainerRow{
						Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
						Children: []components.PageInterface{
							components.ContainerError{
								Error: getters.GetterKey("$error.SmtpUsername"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmtpUsername",
										Label:  "SMTP Username",
										Getter: getters.GetterKey("$in.SmtpUsername"),
									},
								},
							},
							components.ContainerError{
								Error: getters.GetterKey("$error.SmtpPassword"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmtpPassword",
										Label:  "SMTP Password",
										Getter: getters.GetterKey("$in.SmtpPassword"),
									},
								},
							},
						},
					},
					components.ContainerError{
						Error: getters.GetterKey("$error.SmtpFrom"),
						Children: []components.PageInterface{
							components.InputText{
								Name:   "SmtpFrom",
								Label:  "SMTP From Address",
								Getter: getters.GetterKey("$in.SmtpFrom"),
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

		lago.RegistryPage.Patch("users.LoginPage", func(oldPage components.PageInterface) components.PageInterface {
			basePage := oldPage
			if scaffold, ok := basePage.(components.ShellAuthScaffold); ok {
				if len(scaffold.Children) > 0 {
					if col, ok := scaffold.Children[0].(components.ContainerColumn); ok {
						if len(col.Children) > 1 {
							if form, ok := col.Children[1].(components.FormComponent); ok {
								var buttons []components.PageInterface
								if smsEnabled {
									buttons = append(buttons, components.ButtonLink{
										Label: "Login with SMS OTP",
										Link:  getters.GetterStatic("/otp/login/sms/"),
									})
								}
								if emailEnabled {
									buttons = append(buttons, components.ButtonLink{
										Label: "Login with Email OTP",
										Link:  getters.GetterStatic("/otp/login/email/"),
									})
								}
								form.ChildrenAction = append(form.ChildrenAction, components.ContainerColumn{
									Classes: "flex flex-col gap-2 mt-4 items-center border-t border-base-300 pt-4 w-full",
									Children: buttons,
								})
								col.Children[1] = form
								scaffold.Children[0] = col
							}
						}
					}
				}
				return scaffold
			}
			return basePage
		})
		return d
	})
}
