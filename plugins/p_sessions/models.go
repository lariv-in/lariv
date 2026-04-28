package p_sessions

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"gorm.io/gorm"
)

// ClassSession is a scheduled block (calendar-style; optional semester/course like Django events/timetable context).
type ClassSession struct {
	gorm.Model

	Title      string
	Room       string
	StartAt    time.Time
	EndAt      time.Time
	IsActive   bool `gorm:"not null;default:true"`
	SemesterID *uint `gorm:"index"`
	Semester   *p_semesters.Semester `gorm:"foreignKey:SemesterID"`
	CourseID   *uint `gorm:"index"`
	Course     *p_courses.Course `gorm:"foreignKey:CourseID"`
}

func init() {
	lago.OnDBInit("p_sessions.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[ClassSession](d)
		return d
	})
	lago.RegistryAdmin.Register("p_sessions", lago.AdminPanel[ClassSession]{
		SearchField: "Title",
		ListFields:  []string{"Title", "Room", "StartAt", "EndAt", "SemesterID", "CourseID"},
	})
}
