package p_nirmancampus_academicrecords

import (
	"time"

	"github.com/lariv-in/lago/lago"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	sessions "github.com/lariv-in/lago/plugins/p_nirmancampus_sessions"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// AcademicRecordStatusChoices is the canonical ordered list of persisted status -> label (select option order).
var AcademicRecordStatusChoices = []registry.Pair[string, string]{
	{Key: "Not Applied", Value: "Not Applied"},
	{Key: "Applied", Value: "Applied"},
	{Key: "Enrolled", Value: "Enrolled"},
	{Key: "Rejected", Value: "Rejected"},
}

// AcademicRecord links a student to a program term with status and course selections.
// CompulsoryCourses and OptionalCourses are many-to-many relations to Course (same pattern as ProgramStructureUnit).
type AcademicRecord struct {
	gorm.Model

	StudentID              uint                                         `gorm:"notnull;index"`
	Student                p_nirmancampus_students.Student              `gorm:"constraint:OnDelete:CASCADE;foreignKey:StudentID;references:ID"`
	ProgramID              uint                                         `gorm:"not null;index"`
	Program                p_nirmancampus_programs.Program              `gorm:"constraint:OnDelete:RESTRICT;foreignKey:ProgramID;references:ID"`
	SessionID              uint                                         `gorm:"not null;index"`
	AdmissionSession       sessions.AdmissionSession                    `gorm:"constraint:OnDelete:RESTRICT;foreignKey:SessionID;references:ID"`
	ProgramStructureUnitID uint                                         `gorm:"not null;index"`
	ProgramStructureUnit   p_nirmancampus_programs.ProgramStructureUnit `gorm:"constraint:OnDelete:RESTRICT;foreignKey:ProgramStructureUnitID;references:ID"`
	Date                   time.Time                                    `gorm:"type:date"`
	Status                 string                                       `gorm:"type:varchar(50);notnull"`
	CompulsoryCourses      []courses.Course                             `gorm:"many2many:academic_record_compulsory_courses;"`
	OptionalCourses        []courses.Course                             `gorm:"many2many:academic_record_optional_courses;"`
}

func init() {
	lago.OnDBInit("p_nirmancampus_academicrecords.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[AcademicRecord](d)
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_academicrecords", lago.AdminPanel[AcademicRecord]{
		SearchField: "Status",
		ListFields:  []string{"Status", "ProgramStructureUnit.TermNumber", "Date", "AdmissionSession.Name", "Program.Name", "Student.StudentNo", "UpdatedAt"},
		Preload:     []string{"Student", "Program", "AdmissionSession", "ProgramStructureUnit", "CompulsoryCourses", "OptionalCourses"},
	})
}
