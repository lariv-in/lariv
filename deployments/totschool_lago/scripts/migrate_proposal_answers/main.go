package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/registry"
	"gorm.io/gorm"
)

type proposalRow struct {
	ID      uint   `gorm:"column:id"`
	Answers string `gorm:"column:answers"`
}

type legacyQA struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func isAlreadyPairShape(raw string) bool {
	if len(raw) == 0 {
		return false
	}
	var pairs []registry.Pair[string, string]
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&pairs); err != nil {
		fmt.Println("error", err)
		return false
	}
	// If at least one item has a non-empty Key or Value, we consider it migrated.
	for _, p := range pairs {
		if p.Key != "" || p.Value != "" {
			return true
		}
	}

	fmt.Println("pairs", pairs)

	// Could be [] or [{"Key":"","Value":""}] etc; treat as migrated to avoid rewriting noise.
	return false
}

func isLegacyQAShape(raw string) bool {
	if len(raw) == 0 {
		return false
	}
	var legacy []legacyQA
	if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
		return false
	}
	if len(legacy) == 0 {
		return false
	}
	// Consider it legacy if at least one object contains either field.
	for _, item := range legacy {
		if item.Question != "" || item.Answer != "" {
			return true
		}
	}
	return false
}

func convertLegacyToPairs(raw string) (string, bool, error) {
	if len(raw) == 0 {
		return "", false, nil
	}

	var legacy []legacyQA
	if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
		return "", false, err
	}

	out := make([]registry.Pair[string, string], 0, len(legacy))
	for _, item := range legacy {
		out = append(out, registry.Pair[string, string]{Key: item.Question, Value: item.Answer})
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", false, err
	}
	return string(b), true, nil
}

func previewJSON(raw string, max int) string {
	if raw == "" {
		return "<nil>"
	}
	s := strings.TrimSpace(raw)
	if s == "" {
		return "<empty>"
	}
	if max > 0 && len(s) > max {
		return s[:max] + "...(truncated)"
	}
	return s
}

func process(db *gorm.DB, dryRun bool, limit int, sampleUnknown int) error {
	var rows []proposalRow
	q := db.Table("proposals").Select("id, answers").Order("id asc")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rows).Error; err != nil {
		return err
	}

	var scanned, updated, skipped, failed int
	var alreadyPairs, legacyConverted, unknown int
	unknownSamples := 0

	for _, row := range rows {
		scanned++

		// Classification for reporting.
		switch {
		case len(row.Answers) == 0:
			// Treat empty as skipped.
		case isAlreadyPairShape(row.Answers):
			alreadyPairs++
		case isLegacyQAShape(row.Answers):
			legacyConverted++
		default:
			unknown++
			if unknownSamples < sampleUnknown {
				unknownSamples++
				log.Printf("UNKNOWN answers shape id=%d answers=%s", row.ID, previewJSON(row.Answers, 500))
			}
		}

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

	fmt.Printf("scanned=%d updated=%d skipped=%d failed=%d dryRun=%v alreadyPairs=%d legacy=%d unknown=%d unknownSamples=%d\n",
		scanned, updated, skipped, failed, dryRun, alreadyPairs, legacyConverted, unknown, unknownSamples)
	if failed > 0 {
		return fmt.Errorf("%d rows failed", failed)
	}
	return nil
}

func main() {
	var (
		configPath    = flag.String("config", "totschool.toml", "Path to lago config TOML")
		dryRun        = flag.Bool("dry-run", true, "If true, do not write updates")
		limit         = flag.Int("limit", 0, "Max rows to scan (0 = no limit)")
		sampleUnknown = flag.Int("sample-unknown", 5, "Print up to N unknown-shaped rows")
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

	if err := process(db, *dryRun, *limit, *sampleUnknown); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}
