package p_admissions

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_programs"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// AdmissionApplication mirrors admissions.Applicant (Program + semester intake).
type AdmissionApplication struct {
	gorm.Model

	ProgramID     uint `gorm:"not null;index"`
	Program       p_programs.Program
	SemesterID    uint `gorm:"not null;index"`
	Semester      p_semesters.Semester
	ApplicantName string
	Email         string
	Status        string `gorm:"not null"`
	Remarks       string `gorm:"type:text"`

	UserID     *uint `gorm:"index"`
	LinkedUser *p_users.User

	AdhaarNo    string
	DOB         *time.Time
	Gender      string
	Nationality string
	Address     string `gorm:"type:text"`
}

// ApplicationStatusChoices align with typical intake workflow (extend to match legacy DB values).
var ApplicationStatusChoices = []registry.Pair[string, string]{
	{Key: "draft", Value: "Draft"},
	{Key: "submitted", Value: "Submitted"},
	{Key: "under_review", Value: "Under review"},
	{Key: "accepted", Value: "Accepted"},
	{Key: "rejected", Value: "Rejected"},
	{Key: "withdrawn", Value: "Withdrawn"},
}

func init() {
	lago.OnDBInit("p_admissions.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[AdmissionApplication](d)
		return d
	})
	lago.RegistryAdmin.Register("p_admissions", lago.AdminPanel[AdmissionApplication]{
		SearchField: "ApplicantName",
		ListFields:  []string{"ProgramID", "SemesterID", "ApplicantName", "Email", "Status"},
	})
}
