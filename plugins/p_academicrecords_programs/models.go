package p_academicrecords_programs

import (
	"errors"
	"log"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_academicrecords"
	"github.com/lariv-in/lago/p_programs"
	"gorm.io/gorm"
)

// AcademicRecordProgramDetails links an AcademicRecord to a Program.
//
// This is a one-to-one style extension model in Lago (the Django source
// attaches `program` to the AcademicRecord model via `add_to_class`).
type AcademicRecordProgramDetails struct {
	gorm.Model

	AcademicRecordID uint `gorm:"uniqueIndex;notnull"`
	AcademicRecord   p_academicrecords.AcademicRecord `gorm:"constraint:OnDelete:CASCADE;foreignKey:AcademicRecordID;references:ID"`

	ProgramID uint `gorm:"notnull"`
	Program   p_programs.Program `gorm:"constraint:OnDelete:CASCADE;foreignKey:ProgramID;references:ID"`
}

func upsertAcademicRecordProgram(tx *gorm.DB, academicRecordID uint, programID uint) error {
	// Empty selection clears the extension record.
	if programID == 0 {
		return tx.Where("academic_record_id = ?", academicRecordID).
			Delete(&AcademicRecordProgramDetails{}).Error
	}

	var existing AcademicRecordProgramDetails
	err := tx.Where("academic_record_id = ?", academicRecordID).Take(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&AcademicRecordProgramDetails{
				AcademicRecordID: academicRecordID,
				ProgramID:         programID,
			}).Error
		}
		return err
	}

	existing.ProgramID = programID
	return tx.Save(&existing).Error
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&AcademicRecordProgramDetails{}); err != nil {
			log.Panicf("failed to migrate AcademicRecordProgramDetails: %v", err)
		}
		return d
	})
}

