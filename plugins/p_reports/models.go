package p_reports

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// ReportDefinition is a named report template; Frequency matches Django reports.Frequency choices.
type ReportDefinition struct {
	gorm.Model

	Name      string
	Notes     string `gorm:"type:text"`
	Frequency string `gorm:"type:varchar(20);default:'One Time'"` // One Time, Daily, Weekly, Monthy, Yearly (Django spelling)
	ReportAt  *time.Time
}

// ReportFrequencyChoices match Django `reports.Frequency` and model default spelling ("Monthy" legacy).
var ReportFrequencyChoices = []registry.Pair[string, string]{
	{Key: "One Time", Value: "One Time"},
	{Key: "Daily", Value: "Daily"},
	{Key: "Weekly", Value: "Weekly"},
	{Key: "Monthy", Value: "Monthly (legacy key Monthy)"},
	{Key: "Yearly", Value: "Yearly"},
}

func init() {
	lago.OnDBInit("p_reports.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[ReportDefinition](d)
		return d
	})
	lago.RegistryAdmin.Register("p_reports", lago.AdminPanel[ReportDefinition]{
		SearchField: "Name",
		ListFields:  []string{"Name", "Frequency", "ReportAt", "Notes"},
	})
}
