package p_assignments_semesters

import (
	"errors"
	"log"
	"log/slog"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_assignments"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"gorm.io/gorm"
)

// AssignmentSemesterDetails links an Assignment to a Semester (addon-owned extension).
type AssignmentSemesterDetails struct {
	gorm.Model

	AssignmentID uint `gorm:"uniqueIndex;notnull"`
	Assignment   p_assignments.Assignment `gorm:"constraint:OnDelete:CASCADE;foreignKey:AssignmentID;references:ID"`

	SemesterID uint `gorm:"notnull"`
	Semester   p_semesters.Semester `gorm:"constraint:OnDelete:CASCADE;foreignKey:SemesterID;references:ID"`
}

// UpsertAssignmentSemester persists or clears semester linkage for an assignment.
func UpsertAssignmentSemester(tx *gorm.DB, assignmentID, semesterID uint) error {
	if assignmentID == 0 {
		return nil
	}
	if semesterID == 0 {
		return tx.Where("assignment_id = ?", assignmentID).
			Delete(&AssignmentSemesterDetails{}).Error
	}

	var existing AssignmentSemesterDetails
	err := tx.Where("assignment_id = ?", assignmentID).Take(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&AssignmentSemesterDetails{
				AssignmentID: assignmentID,
				SemesterID:   semesterID,
			}).Error
		}
		return err
	}

	existing.SemesterID = semesterID
	return tx.Save(&existing).Error
}

func backfillLegacySemesterColumn(db *gorm.DB) {
	var rows []struct {
		ID         uint `gorm:"column:id"`
		SemesterID uint `gorm:"column:semester_id"`
	}
	if err := db.Table("assignments").
		Select("id", "semester_id").
		Where("semester_id IS NOT NULL AND semester_id != 0").
		Find(&rows).Error; err != nil {
		slog.Info("p_assignments_semesters: skip legacy semester_id backfill", "error", err)
		return
	}
	for _, r := range rows {
		if err := UpsertAssignmentSemester(db, r.ID, r.SemesterID); err != nil {
			slog.Error("p_assignments_semesters: legacy backfill failed", "assignment_id", r.ID, "error", err)
		}
	}
	if len(rows) > 0 {
		slog.Info("p_assignments_semesters: backfilled assignment semester links", "rows", len(rows))
	}
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&AssignmentSemesterDetails{}); err != nil {
			log.Panicf("failed to migrate AssignmentSemesterDetails: %v", err)
		}
		backfillLegacySemesterColumn(d)
		return d
	})
}
