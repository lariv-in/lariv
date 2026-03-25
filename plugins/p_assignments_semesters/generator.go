package p_assignments_semesters

import (
	"errors"
	"fmt"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_assignments"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"gorm.io/gorm"
)

// Generator assignments_semesters.Generator links existing assignments to semesters
// (round-robin). Run after assignments.Generator in deployment GeneratorOrder (create phase).
func init() {
	lago.RegistryGenerator.Register("assignments_semesters.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			var semesters []p_semesters.Semester
			if err := db.Order("id ASC").Find(&semesters).Error; err != nil {
				return fmt.Errorf("failed to load semesters: %w", err)
			}
			if len(semesters) == 0 {
				return fmt.Errorf("need at least one semester before generating assignment-semester links")
			}

			var assignments []p_assignments.Assignment
			if err := db.Order("id ASC").Find(&assignments).Error; err != nil {
				return fmt.Errorf("failed to load assignments: %w", err)
			}
			if len(assignments) == 0 {
				return fmt.Errorf("need at least one assignment before generating assignment-semester links")
			}

			n := 0
			for i, a := range assignments {
				var existing AssignmentSemesterDetails
				err := db.Where("assignment_id = ?", a.ID).Take(&existing).Error
				if err == nil {
					continue
				}
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("lookup extension for assignment %d: %w", a.ID, err)
				}

				sem := semesters[i%len(semesters)]
				if err := UpsertAssignmentSemester(db, a.ID, sem.ID); err != nil {
					return fmt.Errorf("link assignment %d to semester %d: %w", a.ID, sem.ID, err)
				}
				n++
			}

			fmt.Printf("Created %d assignment-semester links (%d assignments, %d semesters)\n", n, len(assignments), len(semesters))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1=1").Delete(&AssignmentSemesterDetails{}).Error
		},
	})
}
