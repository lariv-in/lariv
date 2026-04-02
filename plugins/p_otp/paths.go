package p_otp

import (
	"github.com/lariv-in/lago/lago"
)

func init() {
	_ = lago.RegistryRoute.Register("otp.ForgotPasswordRoute", lago.Route{
		Path:    "/otp/forgot-password/",
		Handler: lago.NewDynamicView("otp.ForgotPasswordView"),
	})

	_ = lago.RegistryRoute.Register("otp.PhoneOtpRequestRoute", lago.Route{
		Path:    "/otp/login/sms/",
		Handler: lago.NewDynamicView("otp.PhoneOtpRequestView"),
	})

	_ = lago.RegistryRoute.Register("otp.EmailOtpRequestRoute", lago.Route{
		Path:    "/otp/login/email/",
		Handler: lago.NewDynamicView("otp.EmailOtpRequestView"),
	})

	_ = lago.RegistryRoute.Register("otp.OtpVerifyRoute", lago.Route{
		Path:    "/otp/verify/",
		Handler: lago.NewDynamicView("otp.OtpVerifyView"),
	})

	_ = lago.RegistryRoute.Register("otp.OTPPreferencesRoute", lago.Route{
		Path:    "/otp/preferences/",
		Handler: lago.NewDynamicView("otp.OTPPreferencesView"),
	})
}
