package p_nirmancampus_studentpayments

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// PaymentMethodChoices is persisted method key -> UI label (slice order = dropdown order).
var PaymentMethodChoices = []registry.Pair[string, string]{
	{Key: "cash", Value: "Cash"},
	{Key: "card", Value: "Card"},
	{Key: "upi", Value: "UPI"},
	{Key: "bank_transfer", Value: "Bank transfer"},
	{Key: "cheque", Value: "Cheque"},
	{Key: "other", Value: "Other"},
}

// Payment is a fee payment recorded for a student.
type Payment struct {
	gorm.Model

	StudentID       uint                            `gorm:"not null;index"`
	Student         p_nirmancampus_students.Student `gorm:"constraint:OnDelete:CASCADE;foreignKey:StudentID;references:ID"`
	Amount          float64                         `gorm:"type:numeric(12,2);not null"`
	PaymentMethod   string                          `gorm:"type:varchar(50);not null"`
	Remarks         string                          `gorm:"type:text"`
	TransactionID   string                          `gorm:"type:varchar(255);default:''"`
	PaidAt          *time.Time                      `gorm:"type:date"`
}

func init() {
	lago.OnDBInit("p_nirmancampus_studentpayments.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Payment](d)
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_studentpayments", lago.AdminPanel[Payment]{
		SearchField: "TransactionID",
		ListFields: []string{
			"Amount",
			"PaymentMethod",
			"TransactionID",
			"PaidAt",
			"Student.StudentNo",
			"Student.Name",
			"UpdatedAt",
		},
		Preload: []string{"Student"},
	})
}
