package p_otp

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

func pluginRoutes() lariv.PluginFeatures[lariv.Route] {
	return lariv.PluginFeatures[lariv.Route]{
		Entries: []registry.Pair[string, lariv.Route]{
			{
				Key: "otp.ForgotPasswordRoute",
				Value: lariv.Route{
					Path:    "/otp/forgot-password/",
					Handler: lariv.NewDynamicView("otp.ForgotPasswordView"),
				},
			},
			{
				Key: "otp.PhoneOtpRequestRoute",
				Value: lariv.Route{
					Path:    "/otp/login/sms/",
					Handler: lariv.NewDynamicView("otp.PhoneOtpRequestView"),
				},
			},
			{
				Key: "otp.EmailOtpRequestRoute",
				Value: lariv.Route{
					Path:    "/otp/login/email/",
					Handler: lariv.NewDynamicView("otp.EmailOtpRequestView"),
				},
			},
			{
				Key: "otp.OtpVerifyRoute",
				Value: lariv.Route{
					Path:    "/otp/verify/",
					Handler: lariv.NewDynamicView("otp.OtpVerifyView"),
				},
			},
			{
				Key: "otp.OTPPreferencesRoute",
				Value: lariv.Route{
					Path:    "/otp/preferences/",
					Handler: lariv.NewDynamicView("otp.OTPPreferencesView"),
				},
			},
		},
	}
}
