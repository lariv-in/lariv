package p_otp

import (
	"fmt"
	"net/http"

	"github.com/lariv-in/components"
	"github.com/lariv-in/lago"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryPage.Register("otp.PhoneOtpRequestForm", components.ShellAuthScaffold{
		Children: []components.PageInterface{
			components.ContainerColumn{
				Classes: "w-80",
				Children: []components.PageInterface{
					components.FieldTitle{Getter: components.GetterStatic("Login via SMS")},
					components.FormComponent{
						Url:    components.GetterStatic("/otp/login/sms/"),
						Method: http.MethodPost,
						ChildrenInput: []components.PageInterface{
							components.InputPhone{
								Name:     "identifier",
								Label:    "Phone Number",
								Required: true,
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
								Link:  components.GetterStatic("/users/login/"),
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
					components.FieldTitle{Getter: components.GetterStatic("Login via Email")},
					components.FormComponent{
						Url:    components.GetterStatic("/otp/login/email/"),
						Method: http.MethodPost,
						ChildrenInput: []components.PageInterface{
							components.InputEmail{
								Name:     "identifier",
								Label:    "Email Address",
								Required: true,
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
								Link:  components.GetterStatic("/users/login/"),
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
					components.FieldTitle{Getter: components.GetterStatic("Verify OTP")},
					components.FormComponent{
						Url:    components.GetterStatic("/otp/verify/"),
						Method: http.MethodPost,
						ChildrenInput: []components.PageInterface{
							components.InputText{
								Name:     "otp",
								Label:    "OTP",
								Required: true,
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
								Link:  components.GetterStatic("/users/login/"),
							},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("otp.OTPPreferencesMenu", components.SidebarMenu{
		Title: components.GetterStatic("OTP Preferences"),
		Back: &components.SidebarMenuItem{
			Title: components.GetterStatic("Back to Home"),
			Url:   components.GetterStatic("/apps/"),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: components.GetterStatic("Preferences"),
				Url:   components.GetterStatic("/otp/preferences/"),
			},
		},
	})

	lago.RegistryPage.Register("otp.OTPPreferencesForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "otp.OTPPreferencesMenu"},
		},
		Children: []components.PageInterface{
			components.FormComponent{
				Url:      components.GetterStatic("/otp/preferences/"),
				Title:    "OTP Preferences",
				Subtitle: "Configure OTP settings for SMS and Email",
				Method:   http.MethodPost,
				ChildrenInput: []components.PageInterface{
					components.FieldTitle{
						Getter: components.GetterStatic("SMS OTP Settings"),
					},
					components.InputText{
						Name:   "msg91_auth_key",
						Label:  "MSG91 Auth Key",
						Getter: components.GetterKey("$in.msg91_auth_key"),
					},
					components.ContainerRow{
						Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
						Children: []components.PageInterface{
							components.InputText{
								Name:   "sms_otp_template_id",
								Label:  "SMS OTP Template ID",
								Getter: components.GetterKey("$in.sms_otp_template_id"),
							},
							components.InputText{
								Name:   "otp_template_id",
								Label:  "General OTP Template ID (Fallback)",
								Getter: components.GetterKey("$in.otp_template_id"),
							},
						},
					},
					components.InputText{
						Name:   "sms_otp_field_name",
						Label:  "SMS OTP Field Name",
						Getter: components.GetterKey("$in.sms_otp_field_name"),
					},
					components.InputText{
						Name:   "sms_otp_extra_fields",
						Label:  "SMS OTP Extra Fields (JSON)",
						Getter: components.GetterKey("$in.sms_otp_extra_fields"),
					},
					components.FieldTitle{
						Getter: components.GetterStatic("Email OTP Settings"),
					},
					components.InputText{
						Name:   "email_otp_template_string",
						Label:  "Email OTP Template String",
						Getter: components.GetterKey("$in.email_otp_template_string"),
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
	lago.OnDbInit(func(d *gorm.DB) *gorm.DB {
		err := lago.RegistryPage.Patch("users.LoginPage", func(oldPage components.PageInterface) components.PageInterface {
			basePage := oldPage
			if scaffold, ok := basePage.(components.ShellAuthScaffold); ok {
				if len(scaffold.Children) > 0 {
					if col, ok := scaffold.Children[0].(components.ContainerColumn); ok {
						if len(col.Children) > 1 {
							if form, ok := col.Children[1].(components.FormComponent); ok {
								form.ChildrenAction = append(form.ChildrenAction, components.ContainerColumn{
									Classes: "flex flex-col gap-2 mt-4 items-center border-t border-base-300 pt-4 w-full",
									Children: []components.PageInterface{
										components.ButtonLink{
											Label: "Login with SMS OTP",
											Link:  components.GetterStatic("/otp/login/sms/"),
										},
										components.ButtonLink{
											Label: "Login with Email OTP",
											Link:  components.GetterStatic("/otp/login/email/"),
										},
									},
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
		if err != nil {
			fmt.Printf("Failed to patch users.LoginPage: %v\n", err)
		}
		return d
	})
}
