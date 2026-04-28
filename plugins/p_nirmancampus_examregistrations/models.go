package p_nirmancampus_examregistrations

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// ExamRegistrationStatusNotRegisteredKey is the default for new rows (create form, bulk create).
const ExamRegistrationStatusNotRegisteredKey = "not_registered"

// ExamRegistrationStatusChoices defines stored keys and UI labels for RegistrationStatus.
var ExamRegistrationStatusChoices = []registry.Pair[string, string]{
	{Key: ExamRegistrationStatusNotRegisteredKey, Value: "Not Registered"},
	{Key: "registered", Value: "Registered"},
}

type ExamRegistration struct {
	gorm.Model

	ExamTitle          string `gorm:"type:varchar(255);not null;index"`
	RegistrationStatus string `gorm:"type:varchar(50);not null;index"`
	Fee                uint   `gorm:"not null;default:0"`

	CourseID uint                          `gorm:"not null;index"`
	Course   p_nirmancampus_courses.Course `gorm:"constraint:OnDelete:RESTRICT;foreignKey:CourseID;references:ID"`

	AcademicRecordID uint                                          `gorm:"not null;index"`
	AcademicRecord   p_nirmancampus_academicrecords.AcademicRecord `gorm:"constraint:OnDelete:CASCADE;foreignKey:AcademicRecordID;references:ID"`

	Assets []p_filesystem.VNode `gorm:"many2many:exam_registration_assets;"`
}

func init() {
	lago.OnDBInit("p_nirmancampus_examregistrations.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[ExamRegistration](d)
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_examregistrations", lago.AdminPanel[ExamRegistration]{
		SearchField: "ExamTitle",
		ListFields: []string{
			"ExamTitle",
			"RegistrationStatus",
			"Fee",
			"Course.Name",
			"AcademicRecord.Student.StudentNo",
			"UpdatedAt",
		},
		Preload: []string{"Course", "AcademicRecord.Student", "Assets"},
	})
}
