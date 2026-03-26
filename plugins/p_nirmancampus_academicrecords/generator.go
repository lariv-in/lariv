package p_nirmancampus_academicrecords

import (
	"fmt"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
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
			var students []p_nirmancampus_students.Student
			if err := db.Order("id ASC").Find(&students).Error; err != nil {
				return fmt.Errorf("failed to load students: %w", err)
			}
			if len(students) == 0 {
				return fmt.Errorf("need at least one student before generating academic records")
			}

			n := 0
			for k, st := range students {
				rec := AcademicRecord{
					StudentID: st.ID,
					Status:    sampleStatuses[k%len(sampleStatuses)],
				}
				if err := db.Create(&rec).Error; err != nil {
					return fmt.Errorf("failed to create academic record (student_id=%d): %w", st.ID, err)
				}
				n++
			}

			fmt.Printf("Created %d academic records (%d students)\n", n, len(students))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1=1").Delete(&AcademicRecord{}).Error
		},
	})
}
