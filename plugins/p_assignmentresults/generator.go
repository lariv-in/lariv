package p_assignmentresults

import (
	"fmt"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/plugins/p_assignments"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryGenerator.Register("assignmentresults.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			var asg p_assignments.Assignment
			if err := db.Order("id ASC").First(&asg).Error; err != nil {
				return fmt.Errorf("assignmentresults generator needs at least one assignment: %w", err)
			}
			var rec p_nirmancampus_academicrecords.AcademicRecord
			if err := db.Order("id ASC").First(&rec).Error; err != nil {
				return fmt.Errorf("assignmentresults generator needs at least one academic record: %w", err)
			}

			var n int64
			if err := db.Model(&AssignmentResult{}).
				Where("assignment_id = ? AND academic_record_id = ?", asg.ID, rec.ID).
				Count(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				fmt.Println("Sample assignment result already exists; skipping")
				return nil
			}

			r := AssignmentResult{
				AssignmentID:     asg.ID,
				AcademicRecordID: rec.ID,
				Marks:            42,
				Remarks:          "Sample generated result.",
			}
			if err := db.Create(&r).Error; err != nil {
				return fmt.Errorf("failed to create sample assignment result: %w", err)
			}
			fmt.Println("Created 1 sample assignment result")
			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1=1").Delete(&AssignmentResult{}).Error
		},
	})
}
