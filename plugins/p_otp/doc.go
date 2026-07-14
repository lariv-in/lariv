// Package p_otp implements one-time password (OTP) delivery and verification features for password recovery and multi-factor auth.
// It integrates SMS gateways (Twilio) and SMTP email servers to send dynamic six-digit OTP codes.
//
// # Registrations and Features Added
//
// # Database Models
//
//   - p_otp.OTPEntity: Caches pending OTP validation codes, timestamps, and target phone/email keys.
//   - p_otp.OTPPreferences: GORM database singleton entity mapping SMTP configurations and Twilio API keys.
//
// # Pages
//
//   - "otp.ForgotPasswordPage" -> components.PageInterface
//         Prompt page allowing choice between SMS or Email recovery options.
//   - "otp.PhoneOtpRequestForm" & "otp.EmailOtpRequestForm": Form panels requesting user identifiers (phone/email).
//   - "otp.OtpVerifyForm": Entry form validating the six-digit code and resetting the user's password.
//   - "otp.OTPPreferencesForm": Admin configuration panel for SMTP/Twilio keys.
//
// # Routes
//
// Registers HTTP ServeMux path mappings:
//
//   - "/otp/forgot-password/" -> otp.ForgotPasswordView
//   - "/otp/login/sms/" -> otp.PhoneOtpRequestView
//   - "/otp/login/email/" -> otp.EmailOtpRequestView
//   - "/otp/verify/" -> otp.OtpVerifyView
//   - "/otp/preferences/" -> otp.OTPPreferencesView
//
// # Views
//
//   - "otp.ForgotPasswordView": Renders choices for SMS vs Email recovery.
//   - "otp.PhoneOtpRequestView" & "otp.EmailOtpRequestView": Handles GET/POST requests to send OTP codes.
//   - "otp.OtpVerifyView": Verifies codes and applies user password resets.
//   - "otp.OTPPreferencesView": Renders and saves SMTP/Twilio configuration settings.
//
// # Patches Applied
//
//   - "p_users.LoginPage": Patches the user login page to insert a "Forgot password?" link directing users to the recovery flow.
package p_otp
