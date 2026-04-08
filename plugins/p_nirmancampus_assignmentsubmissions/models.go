package p_nirmancampus_assignmentsubmissions

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"gorm.io/gorm"
)

type AssignmentSubmission struct {
	gorm.Model

	AssignmentTitle  string `gorm:"type:varchar(255);not null;index"`
	MaxMarks         int    `gorm:"not null"`
	SubmissionStatus string `gorm:"type:varchar(50);not null;index"`
	Marks            int    `gorm:"not null"`

	CourseID uint                          `gorm:"not null;index"`
	Course   p_nirmancampus_courses.Course `gorm:"constraint:OnDelete:RESTRICT;foreignKey:CourseID;references:ID"`

	AcademicRecordID uint                                          `gorm:"not null;index"`
	AcademicRecord   p_nirmancampus_academicrecords.AcademicRecord `gorm:"constraint:OnDelete:CASCADE;foreignKey:AcademicRecordID;references:ID"`

	Assets []p_filesystem.VNode `gorm:"many2many:assignment_submission_assets;"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[AssignmentSubmission](d)
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_assignmentsubmissions", lago.AdminPanel[AssignmentSubmission]{
		SearchField: "AssignmentTitle",
		ListFields: []string{
			"AssignmentTitle",
			"SubmissionStatus",
			"Marks",
			"MaxMarks",
			"Course.Name",
			"AcademicRecord.Student.StudentNo",
			"UpdatedAt",
		},
		Preload: []string{"Course", "AcademicRecord.Student", "Assets"},
	})
}
