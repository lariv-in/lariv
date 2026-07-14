package p_otp

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"
)

func pluginRoutes() lago.PluginFeatures[lago.Route] {
	return lago.PluginFeatures[lago.Route]{
		Entries: []registry.Pair[string, lago.Route]{
			{
				Key: "otp.ForgotPasswordRoute",
				Value: lago.Route{
					Path:    "/otp/forgot-password/",
					Handler: lago.NewDynamicView("otp.ForgotPasswordView"),
				},
			},
			{
				Key: "otp.PhoneOtpRequestRoute",
				Value: lago.Route{
					Path:    "/otp/login/sms/",
					Handler: lago.NewDynamicView("otp.PhoneOtpRequestView"),
				},
			},
			{
				Key: "otp.EmailOtpRequestRoute",
				Value: lago.Route{
					Path:    "/otp/login/email/",
					Handler: lago.NewDynamicView("otp.EmailOtpRequestView"),
				},
			},
			{
				Key: "otp.OtpVerifyRoute",
				Value: lago.Route{
					Path:    "/otp/verify/",
					Handler: lago.NewDynamicView("otp.OtpVerifyView"),
				},
			},
			{
				Key: "otp.OTPPreferencesRoute",
				Value: lago.Route{
					Path:    "/otp/preferences/",
					Handler: lago.NewDynamicView("otp.OTPPreferencesView"),
				},
			},
		},
	}
}
