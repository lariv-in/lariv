package p_assignments

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_semesters"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryGenerator.Register("assignments.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			var semesters []p_semesters.Semester
			if err := db.Order("id ASC").Find(&semesters).Error; err != nil {
				return fmt.Errorf("failed to load semesters: %w", err)
			}
			if len(semesters) == 0 {
				return fmt.Errorf("need at least one semester row before generating assignments")
			}

			now := time.Now()
			items := []Assignment{
				{
					Name:        "Problem Set 1 — Algebra",
					Due:         now.AddDate(0, 0, 14),
					Description: "Complete exercises 1–12 from the course text. Show all working.",
					MaxMarks:    100,
				},
				{
					Name:        "Essay: Industrial Revolution",
					Due:         now.AddDate(0, 0, 21),
					Description: "1500–2000 words; cite at least three sources.",
					MaxMarks:    50,
				},
				{
					Name:        "Lab report: pendulum",
					Due:         now.AddDate(0, 0, 7),
					Description: "Include raw data tables, graph, and error analysis.",
					MaxMarks:    30,
				},
			}
			for i := range items {
				items[i].SemesterID = semesters[i%len(semesters)].ID
			}
			for _, a := range items {
				if err := db.Create(&a).Error; err != nil {
					return fmt.Errorf("failed to create assignment %q: %w", a.Name, err)
				}
			}
			fmt.Printf("Created %d assignments (%d semesters)\n", len(items), len(semesters))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Exec("DELETE FROM assignment_assets").Error; err != nil {
				slog.Error("failed clearing assignment_assets join table", "error", err)
			}
			return db.Unscoped().Where("1=1").Delete(&Assignment{}).Error
		},
	})
}
