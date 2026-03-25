package p_assignmentresults

import (
	"log"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_academicrecords"
	"github.com/lariv-in/lago/plugins/p_assignments"
	"gorm.io/gorm"
)

// AssignmentResult stores marks and remarks for one academic record on one assignment.
type AssignmentResult struct {
	gorm.Model

	AssignmentID uint                     `gorm:"notnull;index;uniqueIndex:idx_assignment_result_pair"`
	Assignment   p_assignments.Assignment `gorm:"constraint:OnDelete:CASCADE;foreignKey:AssignmentID;references:ID"`

	AcademicRecordID uint                             `gorm:"notnull;index;uniqueIndex:idx_assignment_result_pair"`
	AcademicRecord   p_academicrecords.AcademicRecord `gorm:"constraint:OnDelete:CASCADE;foreignKey:AcademicRecordID;references:ID"`

	Marks   int    `gorm:"notnull"`
	Remarks string `gorm:"type:text"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&AssignmentResult{}); err != nil {
			log.Panicf("failed to migrate AssignmentResult model: %v", err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_assignmentresults", lago.AdminPanel[AssignmentResult]{
		SearchField: "Remarks",
		ListFields: []string{
			"Assignment.Name",
			"AcademicRecord.Student.StudentNo",
			"AcademicRecord.Student.User.Name",
			"AcademicRecord.Semester.Name",
			"Marks",
			"Remarks",
			"UpdatedAt",
		},
		Preload: []string{"Assignment", "AcademicRecord", "AcademicRecord.Student", "AcademicRecord.Student.User", "AcademicRecord.Semester"},
	})
}
