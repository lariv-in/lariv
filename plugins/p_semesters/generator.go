package p_semesters

import (
	"fmt"
	"time"

	"github.com/lariv-in/lago"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryGenerator.Register("semesters.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			loc := time.UTC
			now := time.Now().In(loc)
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

			// Current term: started ~1 month ago, ends ~6 months after start; today lies inside.
			currentStart := today.AddDate(0, -1, 0)
			currentEnd := currentStart.AddDate(0, 6, 0)

			// Previous term: ends the day before the current term starts.
			pastEnd := currentStart.AddDate(0, 0, -1)
			pastStart := pastEnd.AddDate(0, -5, 0)

			// Next term: starts the day after the current term ends.
			futureStart := currentEnd.AddDate(0, 0, 1)
			futureEnd := futureStart.AddDate(0, 6, 0)

			rows := []Semester{
				{
					Name:     fmt.Sprintf("Generated previous (%s – %s)", pastStart.Format("Jan 2006"), pastEnd.Format("Jan 2006")),
					Start:    pastStart,
					End:      pastEnd,
					IsActive: false,
				},
				{
					Name:     fmt.Sprintf("Generated current (%s – %s)", currentStart.Format("Jan 2006"), currentEnd.Format("Jan 2006")),
					Start:    currentStart,
					End:      currentEnd,
					IsActive: true,
				},
				{
					Name:     fmt.Sprintf("Generated upcoming (%s – %s)", futureStart.Format("Jan 2006"), futureEnd.Format("Jan 2006")),
					Start:    futureStart,
					End:      futureEnd,
					IsActive: false,
				},
			}

			for i := range rows {
				if err := db.Create(&rows[i]).Error; err != nil {
					return fmt.Errorf("failed to create semester %q: %w", rows[i].Name, err)
				}
			}

			fmt.Printf("Created 3 semesters: previous (id=%d), current (id=%d), upcoming (id=%d)\n",
				rows[0].ID, rows[1].ID, rows[2].ID)
			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1=1").Delete(&Semester{}).Error
		},
	})
}
