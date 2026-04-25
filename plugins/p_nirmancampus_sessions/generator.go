package p_nirmancampus_sessions

import (
	"context"
	"fmt"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryGenerator.Register("sessions.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			loc := time.UTC
			now := time.Now().In(loc)
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

			currentStart := today.AddDate(0, -1, 0)
			currentEnd := currentStart.AddDate(0, 6, 0)
			pastEnd := currentStart.AddDate(0, 0, -1)
			pastStart := pastEnd.AddDate(0, -5, 0)
			futureStart := currentEnd.AddDate(0, 0, 1)
			futureEnd := futureStart.AddDate(0, 6, 0)

			admitRows := []AdmissionSession{
				{
					Name:     fmt.Sprintf("%s – %s", pastStart.Format("Jan 2006"), pastEnd.Format("Jan 2006")),
					Start:    pastStart,
					End:      pastEnd,
					IsActive: false,
				},
				{
					Name:     fmt.Sprintf("%s – %s", currentStart.Format("Jan 2006"), currentEnd.Format("Jan 2006")),
					Start:    currentStart,
					End:      currentEnd,
					IsActive: true,
				},
				{
					Name:     fmt.Sprintf("%s – %s", futureStart.Format("Jan 2006"), futureEnd.Format("Jan 2006")),
					Start:    futureStart,
					End:      futureEnd,
					IsActive: false,
				},
			}

			for i := range admitRows {
				if err := gorm.G[AdmissionSession](db).Create(context.Background(), &admitRows[i]); err != nil {
					return fmt.Errorf("failed to create admission session %q: %w", admitRows[i].Name, err)
				}
			}

			examRows := []ExamSession{
				{
					Name:     fmt.Sprintf("Exam %s", pastStart.Format("Jan 2006")),
					Start:    pastStart,
					End:      pastEnd,
					IsActive: false,
				},
				{
					Name:     fmt.Sprintf("Exam %s", currentStart.Format("Jan 2006")),
					Start:    currentStart,
					End:      currentEnd,
					IsActive: true,
				},
				{
					Name:     fmt.Sprintf("Exam %s", futureStart.Format("Jan 2006")),
					Start:    futureStart,
					End:      futureEnd,
					IsActive: false,
				},
			}

			for i := range examRows {
				if err := gorm.G[ExamSession](db).Create(context.Background(), &examRows[i]); err != nil {
					return fmt.Errorf("failed to create exam session %q: %w", examRows[i].Name, err)
				}
			}

			fmt.Printf("Created 3 admission sessions (ids %d,%d,%d) and 3 exam sessions (ids %d,%d,%d)\n",
				admitRows[0].ID, admitRows[1].ID, admitRows[2].ID,
				examRows[0].ID, examRows[1].ID, examRows[2].ID)
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Unscoped().Where("1=1").Delete(&AdmissionSession{}).Error; err != nil {
				return err
			}
			return db.Unscoped().Where("1=1").Delete(&ExamSession{}).Error
		},
	})
}
