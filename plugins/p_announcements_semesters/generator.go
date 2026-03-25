package p_announcements_semesters

import (
	"errors"
	"fmt"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_announcements"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"gorm.io/gorm"
)

// Generator announcements_semesters.Generator links existing announcements to semesters
// (round-robin). Run after announcements.Generator in deployment GeneratorOrder (create phase).
// Remove phase: list this before announcements.Generator so extension rows are cleared first,
// or rely on CASCADE when announcements are removed (either order works with ON DELETE CASCADE).
func init() {
	lago.RegistryGenerator.Register("announcements_semesters.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			var semesters []p_semesters.Semester
			if err := db.Order("id ASC").Find(&semesters).Error; err != nil {
				return fmt.Errorf("failed to load semesters: %w", err)
			}
			if len(semesters) == 0 {
				return fmt.Errorf("need at least one semester before generating announcement-semester links")
			}

			var announcements []p_announcements.Announcement
			if err := db.Order("id ASC").Find(&announcements).Error; err != nil {
				return fmt.Errorf("failed to load announcements: %w", err)
			}
			if len(announcements) == 0 {
				return fmt.Errorf("need at least one announcement before generating announcement-semester links")
			}

			n := 0
			for i, a := range announcements {
				var existing AnnouncementSemesterDetails
				err := db.Where("announcement_id = ?", a.ID).Take(&existing).Error
				if err == nil {
					continue
				}
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("lookup extension for announcement %d: %w", a.ID, err)
				}

				sem := semesters[i%len(semesters)]
				if err := UpsertAnnouncementSemester(db, a.ID, sem.ID); err != nil {
					return fmt.Errorf("link announcement %d to semester %d: %w", a.ID, sem.ID, err)
				}
				n++
			}

			fmt.Printf("Created %d announcement-semester links (%d announcements, %d semesters)\n", n, len(announcements), len(semesters))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1=1").Delete(&AnnouncementSemesterDetails{}).Error
		},
	})
}
