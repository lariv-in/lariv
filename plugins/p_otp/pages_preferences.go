package p_otp

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerOtpPreferencesPages() {
	lago.RegistryPage.Register("otp.OTPPreferencesMenu", components.SidebarMenu{
		Title: getters.Static("OTP Preferences"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Home"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.Static("Preferences"),
				Url:   lago.RoutePath("otp.OTPPreferencesRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("otp.OTPPreferencesForm", components.ShellScaffold{
		Page: components.Page{Roles: []string{"superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "otp.OTPPreferencesMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("otp.OTPPreferencesForm"),
				ActionURL: lago.RoutePath("otp.OTPPreferencesRoute", nil),
				Children: []components.PageInterface{
					components.FormComponent[OTPPreferences]{
						Attr: getters.FormBubbling(getters.Static("otp.OTPPreferencesForm")),

						Title:    "OTP Preferences",
						Subtitle: "Configure OTP settings for SMS and Email",
						ChildrenInput: []components.PageInterface{
							components.FieldText{
								Classes: "text-lg font-semibold mt-4",
								Getter:  getters.Static("SMS OTP Settings"),
							},
							components.ContainerError{
								Error: getters.Key[error]("$error.Msg91AuthKey"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "Msg91AuthKey",
										Label:  "MSG91 Auth Key",
										Getter: getters.Key[string]("$in.Msg91AuthKey"),
									},
								},
							},
							components.ContainerRow{
								Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
								Children: []components.PageInterface{
									components.ContainerError{
										Error: getters.Key[error]("$error.SmsOtpTemplateId"),
										Children: []components.PageInterface{
											components.InputText{
												Name:   "SmsOtpTemplateId",
												Label:  "SMS OTP Template ID",
												Getter: getters.Key[string]("$in.SmsOtpTemplateId"),
											},
										},
									},
									components.ContainerError{
										Error: getters.Key[error]("$error.OtpTemplateId"),
										Children: []components.PageInterface{
											components.InputText{
												Name:   "OtpTemplateId",
												Label:  "General OTP Template ID (Fallback)",
												Getter: getters.Key[string]("$in.OtpTemplateId"),
											},
										},
									},
								},
							},
							components.ContainerError{
								Error: getters.Key[error]("$error.SmsOtpFieldName"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmsOtpFieldName",
										Label:  "SMS OTP Field Name",
										Getter: getters.Key[string]("$in.SmsOtpFieldName"),
									},
								},
							},
							components.ContainerError{
								Error: getters.Key[error]("$error.SmsOtpExtraFields"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmsOtpExtraFields",
										Label:  "SMS OTP Extra Fields (JSON)",
										Getter: getters.Key[string]("$in.SmsOtpExtraFields"),
									},
								},
							},
							components.FieldText{
								Classes: "text-lg font-semibold mt-4",
								Getter:  getters.Static("Email OTP Settings"),
							},
							components.ContainerError{
								Error: getters.Key[error]("$error.EmailOtpTemplateString"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "EmailOtpTemplateString",
										Label:  "Email OTP Template String",
										Getter: getters.Key[string]("$in.EmailOtpTemplateString"),
									},
								},
							},
							components.ContainerRow{
								Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
								Children: []components.PageInterface{
									components.ContainerError{
										Error: getters.Key[error]("$error.SmtpHost"),
										Children: []components.PageInterface{
											components.InputText{
												Name:   "SmtpHost",
												Label:  "SMTP Host",
												Getter: getters.Key[string]("$in.SmtpHost"),
											},
										},
									},
									components.ContainerError{
										Error: getters.Key[error]("$error.SmtpPort"),
										Children: []components.PageInterface{
											components.InputText{
												Name:   "SmtpPort",
												Label:  "SMTP Port",
												Getter: getters.Key[string]("$in.SmtpPort"),
											},
										},
									},
								},
							},
							components.ContainerRow{
								Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
								Children: []components.PageInterface{
									components.ContainerError{
										Error: getters.Key[error]("$error.SmtpUsername"),
										Children: []components.PageInterface{
											components.InputText{
												Name:   "SmtpUsername",
												Label:  "SMTP Username",
												Getter: getters.Key[string]("$in.SmtpUsername"),
											},
										},
									},
									components.ContainerError{
										Error: getters.Key[error]("$error.SmtpPassword"),
										Children: []components.PageInterface{
											components.InputText{
												Name:   "SmtpPassword",
												Label:  "SMTP Password",
												Getter: getters.Key[string]("$in.SmtpPassword"),
											},
										},
									},
								},
							},
							components.ContainerError{
								Error: getters.Key[error]("$error.SmtpFrom"),
								Children: []components.PageInterface{
									components.InputText{
										Name:   "SmtpFrom",
										Label:  "SMTP From Address",
										Getter: getters.Key[string]("$in.SmtpFrom"),
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
			},
		},
	})
}
