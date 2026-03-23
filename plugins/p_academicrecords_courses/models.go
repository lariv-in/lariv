package p_academicrecords_courses

import (
	"log"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_academicrecords"
	"github.com/lariv-in/lago/p_courses"
	"gorm.io/gorm"
)

// AcademicRecordCourse is the join row for the many-to-many between academic records and courses.
type AcademicRecordCourse struct {
	gorm.Model

	AcademicRecordID uint `gorm:"notnull;index;uniqueIndex:idx_academic_record_course_pair"`
	AcademicRecord   p_academicrecords.AcademicRecord `gorm:"constraint:OnDelete:CASCADE"`

	CourseID uint             `gorm:"notnull;index;uniqueIndex:idx_academic_record_course_pair"`
	Course   p_courses.Course `gorm:"constraint:OnDelete:CASCADE"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&AcademicRecordCourse{}); err != nil {
			log.Panicf("failed to migrate AcademicRecordCourse model: %v", err)
		}
		return d
	})
}
