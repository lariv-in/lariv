package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/registry"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type proposalRow struct {
	ID      uint           `gorm:"column:id"`
	Answers datatypes.JSON `gorm:"column:answers"`
}

type legacyQA struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func isAlreadyPairShape(raw datatypes.JSON) bool {
	if len(raw) == 0 {
		return true
	}
	var pairs []registry.Pair[string, string]
	if err := json.Unmarshal(raw, &pairs); err != nil {
		return false
	}
	// If at least one item has a non-empty Key or Value, we consider it migrated.
	for _, p := range pairs {
		if p.Key != "" || p.Value != "" {
			return true
		}
	}
	// Could be [] or [{"Key":"","Value":""}] etc; treat as migrated to avoid rewriting noise.
	return true
}

func convertLegacyToPairs(raw datatypes.JSON) (datatypes.JSON, bool, error) {
	if len(raw) == 0 {
		return raw, false, nil
	}

	// If it already parses as Pair shape, do nothing.
	if isAlreadyPairShape(raw) {
		return raw, false, nil
	}

	var legacy []legacyQA
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return nil, false, err
	}
	if len(legacy) == 0 {
		return raw, false, nil
	}

	out := make([]registry.Pair[string, string], 0, len(legacy))
	for _, item := range legacy {
		out = append(out, registry.Pair[string, string]{Key: item.Question, Value: item.Answer})
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, false, err
	}
	return datatypes.JSON(b), true, nil
}

func process(db *gorm.DB, dryRun bool, limit int) error {
	var rows []proposalRow
	q := db.Table("proposals").Select("id, answers").Order("id asc")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rows).Error; err != nil {
		return err
	}

	var scanned, updated, skipped, failed int
	for _, row := range rows {
		scanned++
		newJSON, changed, err := convertLegacyToPairs(row.Answers)
		if err != nil {
			failed++
			log.Printf("id=%d: failed to convert answers: %v", row.ID, err)
			continue
		}
		if !changed {
			skipped++
			continue
		}

		if dryRun {
			updated++
			continue
		}

		if err := db.Table("proposals").
			Where("id = ?", row.ID).
			Update("answers", newJSON).Error; err != nil {
			failed++
			log.Printf("id=%d: failed to update answers: %v", row.ID, err)
			continue
		}
		updated++
	}

	fmt.Printf("scanned=%d updated=%d skipped=%d failed=%d dryRun=%v\n", scanned, updated, skipped, failed, dryRun)
	if failed > 0 {
		return fmt.Errorf("%d rows failed", failed)
	}
	return nil
}

func main() {
	var (
		configPath = flag.String("config", "totschool.toml", "Path to lago config TOML")
		dryRun     = flag.Bool("dry-run", true, "If true, do not write updates")
		limit      = flag.Int("limit", 0, "Max rows to scan (0 = no limit)")
	)
	flag.Parse()

	cfg, err := lago.LoadConfigFromFile(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	db, err := lago.InitDB(cfg)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	if err := process(db, *dryRun, *limit); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

