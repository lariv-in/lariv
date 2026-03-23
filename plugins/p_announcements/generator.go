package p_announcements

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_semesters"
	"github.com/lariv-in/lago/p_users"
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
			var semesters []p_semesters.Semester
			if err := db.Order("id ASC").Find(&semesters).Error; err != nil {
				return fmt.Errorf("failed to load semesters: %w", err)
			}
			if len(semesters) == 0 {
				return fmt.Errorf("need at least one semester row before generating announcements")
			}

			createdByID := pickCreatedByForAnnouncements(db)

			const perSemester = 4
			now := time.Now()
			total := 0
			for si, sem := range semesters {
				for i := range perSemester {
					base := announcementTitles[(si*perSemester+i)%len(announcementTitles)]
					title := base

					release := now.Add(-time.Duration(si*perSemester+i) * 24 * time.Hour)
					var expiry *time.Time
					if rand.Intn(100) < 40 {
						t := release.AddDate(0, 0, 14+rand.Intn(60))
						expiry = &t
					}

					a := Announcement{
						Title:       title,
						Description: fmt.Sprintf("This is sample generated content for: %s. Please refer to the office for official documents.", title),
						CreatedByID: createdByID,
						ReleaseAt:   release,
						ExpiryAt:    expiry,
						SemesterID:  sem.ID,
					}
					if err := db.Create(&a).Error; err != nil {
						return fmt.Errorf("failed to create announcement %q: %w", title, err)
					}
					total++
				}
			}

			fmt.Printf("Created %d announcements (%d per semester, %d semesters)\n", total, perSemester, len(semesters))
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
