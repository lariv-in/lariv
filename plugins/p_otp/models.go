package p_otp

import (
	"encoding/json"
	"log/slog"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// OTPPreferences represents the singleton configuration for the p_otp plugin.
type OTPPreferences struct {
	gorm.Model

	SmsOtpTemplateId       string
	OtpTemplateId          string
	Msg91AuthKey           string
	SmsOtpFieldName        string
	SmsOtpExtraFields      string
	EmailOtpTemplateString string
	SmtpHost               string
	SmtpPort               string
	SmtpUsername           string
	SmtpPassword           string
	SmtpFrom               string
}

// GetExtraFields parses the SmsOtpExtraFields JSON string into a map.
func (p *OTPPreferences) GetExtraFields() map[string]any {
	var fields map[string]any
	if p.SmsOtpExtraFields != "" {
		if err := json.Unmarshal([]byte(p.SmsOtpExtraFields), &fields); err != nil {
			slog.Error("failed to unmarshal SmsOtpExtraFields JSON", "err", err, "value", p.SmsOtpExtraFields)
			return map[string]any{}
		}
	} else {
		fields = map[string]any{}
	}
	return fields
}

// LoadPreferences retrieves the singleton OTPPreferences instance, creating it if it doesn't exist.
func LoadPreferences(db *gorm.DB) OTPPreferences {
	var prefs OTPPreferences
	if err := db.FirstOrCreate(&prefs, OTPPreferences{Model: gorm.Model{ID: 1}}).Error; err != nil {
		slog.Warn("Error while loading preference", "err", err)
		// Log error if needed, but return default empty struct or the partially filled struct
	}
	return prefs
}

func init() {
	lago.OnDBInit("p_otp.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[OTPPreferences](d)
		return d
	})

	lago.RegistryAdmin.Register("p_otp", lago.AdminPanel[OTPPreferences]{SearchField: ""})
}
