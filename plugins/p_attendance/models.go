package p_attendance

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_programs"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"github.com/lariv-in/lago/plugins/p_sessions"
	"github.com/lariv-in/lago/plugins/p_students"
	"gorm.io/gorm"
)

// AttendanceMark mirrors attendance.AttendanceRecord (optional dims use NULL, not 0, for FK clarity).
type AttendanceMark struct {
	gorm.Model

	StudentID uint `gorm:"not null;index"`
	Student   p_students.Student
	SessionID *uint `gorm:"index"`
	Session   *p_sessions.ClassSession `gorm:"foreignKey:SessionID"`
	CourseID  *uint `gorm:"index"`
	Course    *p_courses.Course `gorm:"foreignKey:CourseID"`
	BatchID   *uint `gorm:"index"`
	Program   *p_programs.Program `gorm:"foreignKey:BatchID;references:ID"`
	SemesterID *uint `gorm:"index"`
	Semester   *p_semesters.Semester `gorm:"foreignKey:SemesterID"`

	IsPresent  bool      `gorm:"not null;default:true"`
	Notes      string    `gorm:"type:text"`
	RecordedAt time.Time `gorm:"not null"`
}

func init() {
	lago.OnDBInit("p_attendance.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[AttendanceMark](d)
		return d
	})
	lago.RegistryAdmin.Register("p_attendance", lago.AdminPanel[AttendanceMark]{
		SearchField: "Notes",
		ListFields:  []string{"StudentID", "RecordedAt", "IsPresent", "SessionID", "CourseID"},
	})
}
