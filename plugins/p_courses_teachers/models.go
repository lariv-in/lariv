package p_courses_teachers

import (
	"log"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_courses"
	"github.com/lariv-in/lago/p_teachers"
	"gorm.io/gorm"
)

type CourseTeacher struct {
	gorm.Model

	CourseID uint             `gorm:"notnull;index;uniqueIndex:idx_course_teacher_pair"`
	Course   p_courses.Course `gorm:"constraint:OnDelete:CASCADE"`

	TeacherID uint               `gorm:"notnull;index;uniqueIndex:idx_course_teacher_pair"`
	Teacher   p_teachers.Teacher `gorm:"constraint:OnDelete:CASCADE"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&CourseTeacher{}); err != nil {
			log.Panicf("failed to migrate CourseTeacher model: %v", err)
		}
		return d
	})
}
