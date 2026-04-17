package p_lacerate

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

const sourceKindDirectMedia = "direct_media"

type DirectMediaSource struct {
	gorm.Model
	URL      string `gorm:"type:text;not null;default:''"`
	SourceID uint   `gorm:"not null;uniqueIndex"`
	Source   Source `gorm:"foreignKey:SourceID"`
}

func (d DirectMediaSource) Fetch(ctx context.Context, db *gorm.DB, existingDedup map[string]struct{}) ([]Intel, error) {
	asset, err := directMediaFetchRoot(ctx, db, d.URL)
	if err != nil {
		slog.Error("lacerate: direct media source fetch", "error", err, "source_id", d.SourceID, "url", d.URL)
		return nil, err
	}
	archiveState := &directMediaArchiveState{}
	intels, err := directMediaExtractAsset(ctx, db, d.SourceID, existingDedup, archiveState, asset, Config.DirectMedia.MaxArchiveDepth)
	if err != nil {
		slog.Error("lacerate: direct media source extract", "error", err, "source_id", d.SourceID, "url", d.URL)
		return nil, err
	}
	if len(intels) == 0 {
		err := fmt.Errorf("direct media extraction produced no intel")
		slog.Error("lacerate: direct media source empty", "error", err, "source_id", d.SourceID, "url", d.URL)
		return nil, err
	}
	return intels, nil
}

func init() {
	SourceKindMap[sourceKindDirectMedia] = SourceDesc{
		Name:  "DirectMedia",
		Model: DirectMediaSource{},
	}
	if err := RegistrySourceKind.Register(sourceKindDirectMedia, func() SourceInterface { return &DirectMediaSource{} }); err != nil {
		panic(err)
	}
	lago.OnDBInit("p_lacerate.direct_media_source_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[DirectMediaSource](db)
		return db
	})
}
