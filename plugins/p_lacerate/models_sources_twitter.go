package p_lacerate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// migrateTwitterSourcesDropLegacyHandleColumn removes the old NOT NULL `handle` column after the
// model moved to JSON [TwitterSource.Handles]. GORM does not drop columns on AutoMigrate.
func migrateTwitterSourcesDropLegacyHandleColumn(db *gorm.DB) {
	if !db.Migrator().HasTable(&TwitterSource{}) {
		return
	}
	cols, err := db.Migrator().ColumnTypes(&TwitterSource{})
	if err != nil {
		slog.Error("lacerate: twitter_sources column types", "error", err)
		return
	}
	hasLegacy := false
	for _, c := range cols {
		if c.Name() == "handle" {
			hasLegacy = true
			break
		}
	}
	if !hasLegacy {
		return
	}

	switch db.Dialector.Name() {
	case "postgres":
		if err := db.Exec(`
			UPDATE twitter_sources
			SET handles = json_build_array(trim(both from handle))
			WHERE (
				handles IS NULL
				OR trim(both from handles::text) IN ('null', '[]', '""')
			)
			  AND handle IS NOT NULL AND trim(both from handle) <> ''
		`).Error; err != nil {
			slog.Error("lacerate: twitter_sources copy handle into handles", "error", err)
		}
		if err := db.Exec(`ALTER TABLE twitter_sources DROP COLUMN handle`).Error; err != nil {
			slog.Error("lacerate: twitter_sources drop legacy handle column", "error", err)
		}
	case "sqlite":
		if err := db.Exec(`
			UPDATE twitter_sources
			SET handles = json_array(trim(handle))
			WHERE (handles IS NULL OR handles = 'null' OR handles = '[]' OR handles = '""')
			  AND handle IS NOT NULL AND trim(handle) <> ''
		`).Error; err != nil {
			slog.Error("lacerate: twitter_sources copy handle into handles", "error", err)
		}
		if err := db.Exec(`ALTER TABLE twitter_sources DROP COLUMN handle`).Error; err != nil {
			slog.Error("lacerate: twitter_sources drop legacy handle column", "error", err)
		}
	default:
		slog.Warn("lacerate: twitter_sources still has legacy handle column; migrate or drop it manually", "dialect", db.Dialector.Name())
	}
}

func intelDedupHashFromTwitterStableID(stableID string) string {
	id := strings.TrimSpace(stableID)
	if id == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:])
}

func normalizeTwitterHandle(raw string) string {
	s := strings.TrimSpace(strings.TrimPrefix(raw, "@"))
	return strings.TrimSpace(s)
}

// TwitterSource configures ingestion from Twitter / X accounts (see global [lacerateConfig.Twitter]).
type TwitterSource struct {
	gorm.Model
	Handles  datatypes.JSON `gorm:"type:json"`
	SourceID uint           `gorm:"not null;uniqueIndex"`
	Source   Source         `gorm:"foreignKey:SourceID"`
}

func twitterTweetToMarkdown(handle string, tw twitterFetchedTweet) string {
	var b strings.Builder
	txt := strings.TrimSpace(tw.Text)
	if txt != "" {
		b.WriteString(txt)
		b.WriteString("\n\n")
	}
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "- **Handle:** @%s\n", handle)
	if !tw.CreatedAt.IsZero() {
		fmt.Fprintf(&b, "- **Posted:** %s\n", tw.CreatedAt.UTC().Format(time.RFC3339))
	}
	if tw.Permalink != "" {
		fmt.Fprintf(&b, "- **Link:** %s\n", tw.Permalink)
	}
	return strings.TrimSpace(b.String())
}

func (t TwitterSource) Fetch(ctx context.Context, db *gorm.DB) ([]Intel, error) {
	if Config == nil || Config.Twitter.FetchMode == "" {
		err := fmt.Errorf("twitter source: configure [plugins.p_lacerate] twitter.fetchMode in totschool.toml")
		slog.Error("lacerate: twitter source fetch", "error", err, "source_id", t.SourceID)
		return nil, err
	}

	var rawHandles []string
	if len(t.Handles) > 0 {
		if err := json.Unmarshal(t.Handles, &rawHandles); err != nil {
			slog.Error("lacerate: twitter source unmarshal handles", "error", err, "source_id", t.SourceID)
			return nil, err
		}
	}
	hasHandle := false
	for _, h := range rawHandles {
		if normalizeTwitterHandle(h) != "" {
			hasHandle = true
			break
		}
	}
	if !hasHandle {
		err := fmt.Errorf("twitter source: no handles configured")
		slog.Error("lacerate: twitter source fetch", "error", err, "source_id", t.SourceID)
		return nil, err
	}

	var out []Intel
	for _, raw := range rawHandles {
		handle := normalizeTwitterHandle(raw)
		if handle == "" {
			continue
		}
		tweets, err := fetchTweetsForHandle(ctx, handle)
		if err != nil {
			return nil, err
		}

		for _, tw := range tweets {
			dedupe := intelDedupHashFromTwitterStableID(tw.ID)
			if dedupe == "" {
				slog.Warn("lacerate: twitter item missing id, skip", "handle", handle)
				continue
			}
			var n int64
			if err := db.Model(&Intel{}).Where("source_id = ? AND dedup_hash = ?", t.SourceID, dedupe).Count(&n).Error; err != nil {
				slog.Error("lacerate: twitter source dedupe count", "error", err, "source_id", t.SourceID)
				return nil, err
			}
			if n > 0 {
				continue
			}

			var previewID *uint
			if img := strings.TrimSpace(tw.ImageURL); img != "" {
				ref := ""
				if tw.Permalink != "" {
					ref = tw.Permalink
				} else {
					ref = "https://x.com/"
				}
				previewID = persistIntelPreviewImage(ctx, db, tw.ID, img, ref)
			}

			dedupeCopy := dedupe
			i := Intel{
				SourceID:       t.SourceID,
				DedupHash:      &dedupeCopy,
				Content:        twitterTweetToMarkdown(handle, tw),
				PreviewImageID: previewID,
			}
			if err := db.Create(&i).Error; err != nil {
				if lacerateUniqueViolation(err) {
					continue
				}
				slog.Error("lacerate: twitter source create intel", "error", err, "source_id", t.SourceID)
				return nil, err
			}
			out = append(out, i)
		}
	}
	return out, nil
}

func init() {
	SourceKindMap["twitter"] = SourceDesc{
		Name:  "Twitter / X",
		Model: TwitterSource{},
	}
	lago.OnDBInit("p_lacerate.twitter_source_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[TwitterSource](db)
		migrateTwitterSourcesDropLegacyHandleColumn(db)
		return db
	})
}
