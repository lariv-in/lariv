package p_lacerate

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/lago"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// TargetOfInterest is a short, accurate description of an entity the user tracks; embedding enables similarity search when [VLEmbedder] is configured.
type TargetOfInterest struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string `gorm:"type:text"`
	// Embedding matches [IntelEmbeddingDim] and [TargetOfInterest.String]; nullable when no embedder or empty string input.
	Embedding *pgvector.Vector `gorm:"type:vector(1024)"`
}

func (TargetOfInterest) TableName() string { return "targets_of_interest" }

// String returns markdown-shaped text for display and for [VLEmbedder] input (name, description).
func (t *TargetOfInterest) String() string {
	if t == nil {
		return ""
	}
	var b strings.Builder
	if n := strings.TrimSpace(t.Name); n != "" {
		fmt.Fprintf(&b, "# %s\n\n", n)
	}
	if d := strings.TrimSpace(t.Description); d != "" {
		b.WriteString(d)
	}
	return strings.TrimSpace(b.String())
}

// BeforeSave sets [TargetOfInterest.Embedding] from name/description ([prepareTargetOfInterestEmbeddingForSave]).
func (t *TargetOfInterest) BeforeSave(tx *gorm.DB) error {
	if tx.Statement.SkipHooks {
		return nil
	}
	return prepareTargetOfInterestEmbeddingForSave(context.Background(), t)
}

func init() {
	lago.OnDBInit("p_lacerate.target_of_interest_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[TargetOfInterest](db)
		return db
	})
}
