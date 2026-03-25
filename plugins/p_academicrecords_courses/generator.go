package p_academicrecords_courses

import (
	"fmt"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_academicrecords"
	"github.com/lariv-in/lago/plugins/p_courses"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryGenerator.Register("academicrecords_courses.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			var courseList []p_courses.Course
			if err := db.Order("id ASC").Find(&courseList).Error; err != nil {
				return fmt.Errorf("failed to load courses: %w", err)
			}
			if len(courseList) == 0 {
				return fmt.Errorf("need at least one course before generating academic record course links")
			}

			var records []p_academicrecords.AcademicRecord
			if err := db.Order("id ASC").Find(&records).Error; err != nil {
				return fmt.Errorf("failed to load academic records: %w", err)
			}
			if len(records) == 0 {
				return fmt.Errorf("need at least one academic record before generating academic record course links")
			}

			var n int
			for i := range records {
				c := courseList[i%len(courseList)]
				row := AcademicRecordCourse{
					AcademicRecordID: records[i].ID,
					CourseID:         c.ID,
				}
				if err := db.Create(&row).Error; err != nil {
					return fmt.Errorf("failed to create academic_record_course (academic_record_id=%d course_id=%d): %w", records[i].ID, c.ID, err)
				}
				n++
			}

			fmt.Printf("Created %d academic record course links (%d records, cycling %d courses)\n", n, len(records), len(courseList))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1=1").Delete(&AcademicRecordCourse{}).Error
		},
	})
}
