package p_academicrecords

import (
	"fmt"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_semesters"
	"github.com/lariv-in/p_students"
	"gorm.io/gorm"
)

var sampleStatuses = []string{
	"Enrolled",
	"Enrolled",
	"Completed",
	"Probation",
	"Withdrawn",
}

func init() {
	lago.RegistryGenerator.Register("academicrecords.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			var students []p_students.Student
			if err := db.Order("id ASC").Find(&students).Error; err != nil {
				return fmt.Errorf("failed to load students: %w", err)
			}
			if len(students) == 0 {
				return fmt.Errorf("need at least one student before generating academic records")
			}

			var semesters []p_semesters.Semester
			if err := db.Order("id ASC").Find(&semesters).Error; err != nil {
				return fmt.Errorf("failed to load semesters: %w", err)
			}
			if len(semesters) == 0 {
				return fmt.Errorf("need at least one semester before generating academic records")
			}

			n := 0
			k := 0
			for _, st := range students {
				for _, sem := range semesters {
					rec := AcademicRecord{
						StudentID:  st.ID,
						SemesterID: sem.ID,
						Status:     sampleStatuses[k%len(sampleStatuses)],
					}
					k++
					if err := db.Create(&rec).Error; err != nil {
						return fmt.Errorf("failed to create academic record (student_id=%d semester_id=%d): %w", st.ID, sem.ID, err)
					}
					n++
				}
			}

			fmt.Printf("Created %d academic records (%d students × %d semesters)\n", n, len(students), len(semesters))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1=1").Delete(&AcademicRecord{}).Error
		},
	})
}
