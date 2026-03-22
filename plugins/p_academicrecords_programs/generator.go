package p_academicrecords_programs

import (
	"fmt"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_academicrecords"
	"github.com/lariv-in/p_programs"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryGenerator.Register("academicrecords_programs.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			var programs []p_programs.Program
			if err := db.Order("id ASC").Find(&programs).Error; err != nil {
				return fmt.Errorf("failed to load programs: %w", err)
			}
			if len(programs) == 0 {
				return fmt.Errorf("need at least one program before generating academic record program details")
			}

			var records []p_academicrecords.AcademicRecord
			if err := db.Order("id ASC").Find(&records).Error; err != nil {
				return fmt.Errorf("failed to load academic records: %w", err)
			}
			if len(records) == 0 {
				return fmt.Errorf("need at least one academic record before generating program details")
			}

			for i := range records {
				prog := programs[i%len(programs)]
				row := AcademicRecordProgramDetails{
					AcademicRecordID: records[i].ID,
					ProgramID:        prog.ID,
				}
				if err := db.Create(&row).Error; err != nil {
					return fmt.Errorf("failed to create program details for academic_record_id=%d: %w", records[i].ID, err)
				}
			}

			fmt.Printf("Created %d academic record program details (cycling %d programs)\n",
				len(records), len(programs))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1=1").Delete(&AcademicRecordProgramDetails{}).Error
		},
	})
}
