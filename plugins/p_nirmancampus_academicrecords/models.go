package p_nirmancampus_academicrecords

import (
	"github.com/lariv-in/lago/lago"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	sessions "github.com/lariv-in/lago/plugins/p_nirmancampus_sessions"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"gorm.io/gorm"
)

// AcademicRecordStatusChoices maps stored AcademicRecord.Status values to display labels.
var AcademicRecordStatusChoices = map[string]string{
	"Enrolled":   "Enrolled",
	"Completed":  "Completed",
	"Probation":  "Probation",
	"Withdrawn":  "Withdrawn",
}

// AcademicRecord links a student to a program term with status and course selections.
// CompulsoryCourses and OptionalCourses are many-to-many relations to Course (same pattern as ProgramStructureUnit).
type AcademicRecord struct {
	gorm.Model

	StudentID         uint                            `gorm:"notnull;index"`
	Student           p_nirmancampus_students.Student `gorm:"constraint:OnDelete:CASCADE;foreignKey:StudentID;references:ID"`
	ProgramID         uint                            `gorm:"not null;index"`
	Program           p_nirmancampus_programs.Program `gorm:"constraint:OnDelete:RESTRICT;foreignKey:ProgramID;references:ID"`
	SessionID         uint                            `gorm:"not null;index"`
	Session           sessions.Session                `gorm:"constraint:OnDelete:RESTRICT;foreignKey:SessionID;references:ID"`
	Term              uint                            `gorm:"not null;index"`
	Status            string                          `gorm:"type:varchar(50);notnull"`
	CompulsoryCourses []courses.Course                `gorm:"many2many:academic_record_compulsory_courses;"`
	OptionalCourses   []courses.Course                `gorm:"many2many:academic_record_optional_courses;"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[AcademicRecord](d)
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_academicrecords", lago.AdminPanel[AcademicRecord]{
		SearchField: "Status",
		ListFields:  []string{"Status", "Term", "Session.Name", "Program.Name", "Student.StudentNo", "UpdatedAt"},
		Preload:     []string{"Student", "Program", "Session", "CompulsoryCourses", "OptionalCourses"},
	})
}
