// Package p_announcements registers announcements.Generator for sample announcements (no semester FK).
// With p_announcements_semesters, list announcements_semesters.Generator after this in GeneratorOrder.
// Remove clears announcement_assets then deletes announcements (extension rows cascade).
package p_announcements

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

var announcementTitles = []string{
	"Mid-term examination schedule",
	"Holiday notice: national holiday",
	"Parent–teacher meeting",
	"Sports day registration open",
	"Library hours update",
	"Fee payment deadline reminder",
	"Workshop on study skills",
	"Campus maintenance window",
	"Scholarship application deadline",
	"Annual day rehearsal schedule",
	"Transport route changes",
	"Cafeteria menu update",
}

func pickCreatedByForAnnouncements(db *gorm.DB) *uint {
	var u p_users.User
	err := db.Where("is_superuser = ?", true).Order("id ASC").First(&u).Error
	if err != nil {
		err = db.Order("id ASC").First(&u).Error
		if err != nil {
			return nil
		}
	}
	id := u.ID
	return &id
}

func init() {
	lago.RegistryGenerator.Register("announcements.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			createdByID := pickCreatedByForAnnouncements(db)

			const count = 16
			now := time.Now()
			for i := range count {
				base := announcementTitles[i%len(announcementTitles)]
				title := base

				release := now.Add(-time.Duration(i) * 24 * time.Hour)
				var expiry *time.Time
				if rand.Intn(100) < 40 {
					t := release.AddDate(0, 0, 14+rand.Intn(60))
					expiry = &t
				}

				var url string
			if rand.Intn(100) < 50 {
				url = fmt.Sprintf("https://example.com/announcements/%d", i+1)
			}

			a := Announcement{
					Title:       title,
					Description: fmt.Sprintf("This is sample generated content for: %s. Please refer to the office for official documents.", title),
					URL:         url,
					CreatedByID: createdByID,
					ReleaseAt:   release,
					ExpiryAt:    expiry,
				}
				if err := db.Create(&a).Error; err != nil {
					return fmt.Errorf("failed to create announcement %q: %w", title, err)
				}
			}

			fmt.Printf("Created %d announcements\n", count)
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Exec("DELETE FROM announcement_assets").Error; err != nil {
				slog.Error("failed clearing announcement_assets join table", "error", err)
			}
			return db.Unscoped().Where("1=1").Delete(&Announcement{}).Error
		},
	})
}
