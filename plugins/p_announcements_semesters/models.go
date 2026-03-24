package p_announcements_semesters

import (
	"errors"
	"log"
	"log/slog"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_announcements"
	"github.com/lariv-in/lago/p_semesters"
	"gorm.io/gorm"
)

// AnnouncementSemesterDetails links an Announcement to a Semester (addon-owned extension).
type AnnouncementSemesterDetails struct {
	gorm.Model

	AnnouncementID uint `gorm:"uniqueIndex;notnull"`
	Announcement   p_announcements.Announcement `gorm:"constraint:OnDelete:CASCADE;foreignKey:AnnouncementID;references:ID"`

	SemesterID uint `gorm:"notnull"`
	Semester   p_semesters.Semester `gorm:"constraint:OnDelete:CASCADE;foreignKey:SemesterID;references:ID"`
}

// UpsertAnnouncementSemester persists or clears semester linkage for an announcement.
func UpsertAnnouncementSemester(tx *gorm.DB, announcementID, semesterID uint) error {
	if announcementID == 0 {
		return nil
	}
	if semesterID == 0 {
		return tx.Where("announcement_id = ?", announcementID).
			Delete(&AnnouncementSemesterDetails{}).Error
	}

	var existing AnnouncementSemesterDetails
	err := tx.Where("announcement_id = ?", announcementID).Take(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&AnnouncementSemesterDetails{
				AnnouncementID: announcementID,
				SemesterID:     semesterID,
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
	if err := db.Table("announcements").
		Select("id", "semester_id").
		Where("semester_id IS NOT NULL AND semester_id != 0").
		Find(&rows).Error; err != nil {
		slog.Info("p_announcements_semesters: skip legacy semester_id backfill", "error", err)
		return
	}
	for _, r := range rows {
		if err := UpsertAnnouncementSemester(db, r.ID, r.SemesterID); err != nil {
			slog.Error("p_announcements_semesters: legacy backfill failed", "announcement_id", r.ID, "error", err)
		}
	}
	if len(rows) > 0 {
		slog.Info("p_announcements_semesters: backfilled announcement semester links", "rows", len(rows))
	}
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&AnnouncementSemesterDetails{}); err != nil {
			log.Panicf("failed to migrate AnnouncementSemesterDetails: %v", err)
		}
		backfillLegacySemesterColumn(d)
		return d
	})
}
