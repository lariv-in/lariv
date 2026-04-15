package p_lacerate

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

const websiteSourceDefaultDepth = 2

type WebsiteSource struct {
	gorm.Model
	URL      string `gorm:"type:text;not null;default:''"`
	SourceID uint   `gorm:"not null;uniqueIndex"`
	Source   Source `gorm:"foreignKey:SourceID"`
}

func (w WebsiteSource) Fetch(ctx context.Context, db *gorm.DB, existingDedup map[string]struct{}) ([]Intel, error) {
	intels, err := NewWebsiteFetchers(ctx).FetchWebsite(w.URL, websiteSourceDefaultDepth)
	if err != nil {
		slog.Error("lacerate: website source fetch", "error", err, "source_id", w.SourceID, "url", w.URL)
		return nil, err
	}
	if len(intels) == 0 {
		return nil, nil
	}
	out := make([]Intel, 0, len(intels))
	for i := range intels {
		dedup := ""
		if intels[i].DedupHash != nil {
			dedup = *intels[i].DedupHash
		}
		if dedup == "" {
			continue
		}
		if _, dup := existingDedup[dedup]; dup {
			continue
		}
		sourceID := w.SourceID
		intels[i].SourceID = &sourceID
		existingDedup[dedup] = struct{}{}
		out = append(out, intels[i])
	}
	return out, nil
}

func init() {
	SourceKindMap["website"] = SourceDesc{
		Name:  "Website",
		Model: WebsiteSource{},
	}
	if err := RegistrySourceKind.Register("website", func() SourceInterface { return &WebsiteSource{} }); err != nil {
		panic(err)
	}
	lago.OnDBInit("p_lacerate.website_source_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[WebsiteSource](db)
		return db
	})
}
