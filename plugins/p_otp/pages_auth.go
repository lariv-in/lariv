package p_otp

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func pageEntriesOtpAuth() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		{Key: "otp.PhoneOtpRequestForm", Value: components.ShellAuthScaffold{
			Children: []components.PageInterface{
				components.ContainerColumn{
					Children: []components.PageInterface{
						components.FieldTitle{Getter: getters.Static("Login via SMS")},
						&components.FormListenBoostedPost{
							Name:      getters.Static("otp.PhoneOtpRequestForm"),
							ActionURL: lariv.RoutePath("otp.PhoneOtpRequestRoute", nil),
							Children: []components.PageInterface{
								components.FormComponent[map[string]string]{
									Attr: getters.FormBubbling(getters.Static("otp.PhoneOtpRequestForm")),
									ChildrenInput: []components.PageInterface{
										components.ContainerError{
											Error: getters.Key[error]("$error.Identifier"),
											Children: []components.PageInterface{
												components.InputPhone{
													Name:     "Identifier",
													Label:    "Phone Number",
													Required: true,
													Getter:   getters.Key[string]("$in.Identifier"),
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
							},
						},
						components.ContainerRow{
							Classes: "text-center mt-4",
							Children: []components.PageInterface{
								components.ButtonLink{
									Label: getters.Static("Back to Login"),
									Link:  lariv.RoutePath("p_users.LoginRoute", nil),
								},
							},
						},
					},
				},
			},
		}},

		{Key: "otp.EmailOtpRequestForm", Value: components.ShellAuthScaffold{
			Children: []components.PageInterface{
				components.ContainerColumn{
					Classes: "w-80",
					Children: []components.PageInterface{
						components.FieldTitle{Getter: getters.Static("Login via Email")},
						&components.FormListenBoostedPost{
							Name:      getters.Static("otp.EmailOtpRequestForm"),
							ActionURL: lariv.RoutePath("otp.EmailOtpRequestRoute", nil),
							Children: []components.PageInterface{
								components.FormComponent[map[string]string]{
									Attr: getters.FormBubbling(getters.Static("otp.EmailOtpRequestForm")),
									ChildrenInput: []components.PageInterface{
										components.ContainerError{
											Error: getters.Key[error]("$error.Identifier"),
											Children: []components.PageInterface{
												components.InputEmail{
													Name:     "Identifier",
													Label:    "Email Address",
													Required: true,
													Getter:   getters.Key[string]("$in.Identifier"),
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
							},
						},
						components.ContainerRow{
							Classes: "text-center mt-4",
							Children: []components.PageInterface{
								components.ButtonLink{
									Label: getters.Static("Back to Login"),
									Link:  lariv.RoutePath("p_users.LoginRoute", nil),
								},
							},
						},
					},
				},
			},
		}},

		{Key: "otp.OtpVerifyForm", Value: components.ShellAuthScaffold{
			Children: []components.PageInterface{
				components.ContainerColumn{
					Classes: "w-80",
					Children: []components.PageInterface{
						components.FieldTitle{Getter: getters.Static("Verify OTP")},
						components.FieldText{
							Classes: "text-sm text-gray-600 mb-2",
							Getter:  getters.Static("Enter the code we sent and choose a new password."),
						},
						&components.FormListenBoostedPost{
							Name:      getters.Static("otp.OtpVerifyForm"),
							ActionURL: getters.Format("%v?identifier=%v", getters.Any(lariv.RoutePath("otp.OtpVerifyRoute", nil)), getters.Any(getters.QueryEscape(getters.Key[string]("$in.Identifier")))),
							Children: []components.PageInterface{
								components.FormComponent[map[string]string]{
									Attr: getters.FormBubbling(getters.Static("otp.OtpVerifyForm")),
									ChildrenInput: []components.PageInterface{
										components.ContainerError{
											Error: getters.Key[error]("$error.Otp"),
											Children: []components.PageInterface{
												components.InputText{
													Name:     "Otp",
													Label:    "OTP",
													Required: true,
													Getter:   getters.Key[string]("$in.Otp"),
												},
											},
										},
										components.ContainerError{
											Error: getters.Key[error]("$error.NewPassword"),
											Children: []components.PageInterface{
												components.InputPassword{
													Name:     "NewPassword",
													Label:    "New password",
													Required: true,
													Getter:   getters.Key[string]("$in.NewPassword"),
												},
											},
										},
										components.ContainerError{
											Error: getters.Key[error]("$error.NewPassword2"),
											Children: []components.PageInterface{
												components.InputPassword{
													Name:     "NewPassword2",
													Label:    "Confirm new password",
													Required: true,
													Getter:   getters.Key[string]("$in.NewPassword2"),
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
							},
						},
						components.ContainerRow{
							Classes: "text-center mt-4",
							Children: []components.PageInterface{
								components.ButtonLink{
									Label: getters.Static("Cancel"),
									Link:  lariv.RoutePath("p_users.LoginRoute", nil),
								},
							},
						},
					},
				},
			},
		}},

		{Key: "otp.ForgotPasswordPage", Value: components.ShellAuthScaffold{
			Children: []components.PageInterface{
				components.ContainerColumn{
					Classes: "w-80",
					Children: []components.PageInterface{
						components.ContainerRow{
							Classes: "items-center",
							Children: []components.PageInterface{
								components.ButtonLink{
									Icon:    "arrow-left",
									Link:    lariv.RoutePath("p_users.LoginRoute", nil),
									Classes: "btn-ghost btn-square",
								},
								components.FieldTitle{
									Getter:  getters.Static("Forgot Password"),
									Classes: "grow text-center",
								},
								components.ButtonLink{
									Icon:    "arrow-left",
									Classes: "btn-ghost btn-square invisible",
								},
							},
						},
						components.ContainerColumn{
							Classes: "gap-2 mt-3",
							Children: []components.PageInterface{
								components.ButtonLink{
									Label:   getters.Static("Reset password with email"),
									Link:    lariv.RoutePath("otp.EmailOtpRequestRoute", nil),
									Classes: "w-full",
								},
								components.ButtonLink{
									Label:   getters.Static("Reset password with phone number"),
									Link:    lariv.RoutePath("otp.PhoneOtpRequestRoute", nil),
									Classes: "w-full",
								},
							},
						},
					},
				},
			},
		}},
	}
}
