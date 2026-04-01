package p_nirmancampus_academicrecords

import (
	"log"

	"github.com/lariv-in/lago/lago"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// Stored AcademicRecord.Status values (use these everywhere so forms and DB stay aligned).
const (
	AcademicRecordStatusEnrolled  = "Enrolled"
	AcademicRecordStatusCompleted = "Completed"
	AcademicRecordStatusProbation = "Probation"
	AcademicRecordStatusWithdrawn = "Withdrawn"
)

// AcademicRecordStatusChoices is the canonical list for select inputs and filters (value = label).
func AcademicRecordStatusChoices() []registry.Pair[string, string] {
	return []registry.Pair[string, string]{
		{Key: AcademicRecordStatusEnrolled, Value: "Enrolled"},
		{Key: AcademicRecordStatusCompleted, Value: "Completed"},
		{Key: AcademicRecordStatusProbation, Value: "Probation"},
		{Key: AcademicRecordStatusWithdrawn, Value: "Withdrawn"},
	}
}

// AcademicRecord links a student to a program term with status and course selections.
// CompulsoryCourses and OptionalCourses are many-to-many relations to Course (same pattern as ProgramStructureUnit).
type AcademicRecord struct {
	gorm.Model

	StudentID         uint                            `gorm:"notnull;index"`
	Student           p_nirmancampus_students.Student `gorm:"constraint:OnDelete:CASCADE;foreignKey:StudentID;references:ID"`
	ProgramID         uint                            `gorm:"not null;index"`
	Program           p_nirmancampus_programs.Program `gorm:"constraint:OnDelete:RESTRICT;foreignKey:ProgramID;references:ID"`
	Term              uint                            `gorm:"not null;index"`
	Status            string                          `gorm:"type:varchar(50);notnull"`
	CompulsoryCourses []courses.Course                `gorm:"many2many:academic_record_compulsory_courses;"`
	OptionalCourses   []courses.Course                `gorm:"many2many:academic_record_optional_courses;"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&AcademicRecord{}); err != nil {
			log.Panicf("failed to migrate AcademicRecord: %v", err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_academicrecords", lago.AdminPanel[AcademicRecord]{
		SearchField: "Status",
		ListFields:  []string{"Status", "Term", "Program.Name", "Student.StudentNo", "UpdatedAt"},
		Preload:     []string{"Student.User", "Program", "CompulsoryCourses", "OptionalCourses"},
	})
}
