package p_courses

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_programs"
	"github.com/lariv-in/lago/plugins/p_students"
	"gorm.io/gorm"
)

// Course mirrors allocation.Course (semester-scoped code in Django; JoinCode UUID).
// Programs / Students = Django M2M batches and students on Course.
type Course struct {
	gorm.Model

	Code        string `gorm:"not null;uniqueIndex:ux_course_semester_code"`
	Name        string `gorm:"not null"`
	IsActive    bool   `gorm:"not null;default:true"`
	Description string `gorm:"type:text"`
	Subject     string
	CourseGroup string
	Remarks     string `gorm:"type:text"`
	SemesterID  *uint  `gorm:"index;uniqueIndex:ux_course_semester_code"`
	JoinCode    string `gorm:"index"` // Django UUID; set on create when mirroring legacy

	Programs []p_programs.Program `gorm:"many2many:course_programs;"`
	Students []p_students.Student  `gorm:"many2many:course_students;"`
}

func init() {
	lago.OnDBInit("p_courses.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Course](d)
		return d
	})
	lago.RegistryAdmin.Register("p_courses", lago.AdminPanel[Course]{
		SearchField: "Name",
		ListFields:  []string{"Code", "Name", "IsActive", "Subject", "SemesterID"},
	})
}
