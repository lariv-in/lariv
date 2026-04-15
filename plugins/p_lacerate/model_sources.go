package p_lacerate

import (
	"context"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// SourceInterface is implemented per [Source.Kind]. Fetch must only build [Intel] rows; persistence runs in [runSourceFetch].
// existingDedup holds dedup_hash values already in DB for this source plus hashes appended during this run (mutated by Fetch).
type SourceInterface interface {
	Fetch(ctx context.Context, db *gorm.DB, existingDedup map[string]struct{}) ([]Intel, error)
}

type SourceDesc struct {
	Name  string
	Model SourceInterface
}

var SourceKindMap = map[string]SourceDesc{}

var SourceKindChoices = []registry.Pair[string, string]{
	{Key: "reddit", Value: "Reddit"},
	{Key: "twitter", Value: "Twitter / X"},
	{Key: "website", Value: "Website"},
}

// RegistrySourceKind holds a constructor per [Source.Kind] that returns a new row value (e.g. &RedditSource{})
// for GORM to scan into; the dynamic type must implement [SourceInterface].
var RegistrySourceKind = registry.NewRegistry[func() SourceInterface]()

type Source struct {
	gorm.Model
	Name     string
	Kind     string
	Duration time.Duration
}

// AfterDelete stops the background fetch worker for this source row.
func (s *Source) AfterDelete(tx *gorm.DB) error {
	if s != nil && s.ID != 0 {
		StopSourceWorker(s.ID)
	}
	return nil
}

// deleteSourceKindExtensionRows removes every per-kind row tied to this [Source] (reddit, twitter, …).
// Call inside a transaction before deleting the [Source] row.
func deleteSourceKindExtensionRows(tx *gorm.DB, sourceID uint) error {
	if sourceID == 0 {
		return nil
	}
	if err := tx.Where("source_id = ?", sourceID).Delete(&RedditSource{}).Error; err != nil {
		return err
	}
	if err := tx.Where("source_id = ?", sourceID).Delete(&TwitterSource{}).Error; err != nil {
		return err
	}
	if err := tx.Where("source_id = ?", sourceID).Delete(&WebsiteSource{}).Error; err != nil {
		return err
	}
	return nil
}

func init() {
	lago.OnDBInit("p_lacerate.source_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[Source](db)
		return db
	})
}
