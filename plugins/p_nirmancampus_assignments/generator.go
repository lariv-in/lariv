// Package p_nirmancampus_assignments registers assignments.Generator for sample assignments (no semester FK).
package p_nirmancampus_assignments

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryGenerator.Register("assignments.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
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
			for _, a := range items {
				if err := db.Create(&a).Error; err != nil {
					return fmt.Errorf("failed to create assignment %q: %w", a.Name, err)
				}
			}
			fmt.Printf("Created %d assignments\n", len(items))
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
