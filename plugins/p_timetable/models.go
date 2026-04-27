package p_timetable

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"gorm.io/gorm"
)

// TimetableSlot mirrors timetable.Period-style weekly slot (DayOfWeek + minutes; Semester optional).
type TimetableSlot struct {
	gorm.Model

	DayOfWeek   uint `gorm:"not null"`
	StartMinute uint `gorm:"not null"`
	EndMinute   uint `gorm:"not null"`
	Label       string
	CourseID    *uint `gorm:"index"`
	Course      *p_courses.Course `gorm:"foreignKey:CourseID"`
	SemesterID  *uint `gorm:"index"`
	Semester    *p_semesters.Semester `gorm:"foreignKey:SemesterID"`
}

func init() {
	lago.OnDBInit("p_timetable.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[TimetableSlot](d)
		return d
	})
	lago.RegistryAdmin.Register("p_timetable", lago.AdminPanel[TimetableSlot]{
		SearchField: "Label",
		ListFields:  []string{"DayOfWeek", "StartMinute", "EndMinute", "Label", "CourseID", "SemesterID"},
	})
}
