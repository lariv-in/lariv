package p_courses_teachers

import (
	"fmt"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_courses"
	"github.com/lariv-in/lago/p_teachers"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryGenerator.Register("courses_teachers.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			var courses []p_courses.Course
			if err := db.Order("id ASC").Find(&courses).Error; err != nil {
				return fmt.Errorf("failed to load courses: %w", err)
			}
			if len(courses) == 0 {
				return fmt.Errorf("need at least one course before generating course-teacher links")
			}

			var teachers []p_teachers.Teacher
			if err := db.Order("id ASC").Find(&teachers).Error; err != nil {
				return fmt.Errorf("failed to load teachers: %w", err)
			}
			if len(teachers) == 0 {
				return fmt.Errorf("need at least one teacher before generating course-teacher links")
			}

			n := 0
			for i := range courses {
				t := teachers[i%len(teachers)]
				row := CourseTeacher{
					CourseID:  courses[i].ID,
					TeacherID: t.ID,
				}
				if err := db.Create(&row).Error; err != nil {
					return fmt.Errorf("failed to create course_teacher (course_id=%d teacher_id=%d): %w", courses[i].ID, t.ID, err)
				}
				n++
			}

			fmt.Printf("Created %d course-teacher links (%d courses, cycling %d teachers)\n", n, len(courses), len(teachers))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1=1").Delete(&CourseTeacher{}).Error
		},
	})
}
