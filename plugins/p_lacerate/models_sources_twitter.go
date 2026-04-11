package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

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

func (t TwitterSource) Fetch(ctx context.Context, db *gorm.DB, existingDedup map[string]struct{}) ([]Intel, error) {
	if Config == nil || Config.Twitter.FetchMode == "" {
		err := fmt.Errorf("twitter source: configure [plugins.p_lacerate] twitter.fetchMode in totschool.toml")
		slog.Error("lacerate: twitter source fetch", "error", err, "source_id", t.SourceID)
		return nil, err
	}

	var handles []string
	if len(t.Handles) > 0 {
		if err := json.Unmarshal(t.Handles, &handles); err != nil {
			slog.Error("lacerate: twitter source unmarshal handles", "error", err, "source_id", t.SourceID)
			return nil, err
		}
	}

	var out []Intel
	for _, handle := range handles {
		handle = strings.TrimSpace(handle)
		if handle == "" {
			continue
		}
		tweets, err := fetchTweetsForHandle(ctx, handle)
		if err != nil {
			return nil, err
		}

		for _, tw := range tweets {
			dedupe := tw.IntelDedupHash()
			if dedupe == "" {
				slog.Warn("lacerate: twitter item missing id, skip", "handle", handle)
				continue
			}
			if _, dup := existingDedup[dedupe]; dup {
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
			existingDedup[dedupe] = struct{}{}
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
	if err := RegistrySourceKind.Register("twitter", func() SourceInterface { return &TwitterSource{} }); err != nil {
		panic(err)
	}
	lago.OnDBInit("p_lacerate.twitter_source_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[TwitterSource](db)
		return db
	})
}
