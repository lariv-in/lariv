package p_seer_intel

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// SeerIntelEmbeddingDim is the pgvector width for [Intel.Embedding]. Kept aligned with the
// embedding model dimension used when generation is implemented (see p_lacerate for reference).
const SeerIntelEmbeddingDim = 1024

// Intel is the canonical Seer intel row: display metadata plus an optional embedding vector.
type Intel struct {
	gorm.Model

	Title   string `gorm:"not null;default:''"`
	Summary string `gorm:"type:text;not null;default:''"`
	// Datetime is the canonical time for this intel item (ingest or event time).
	Datetime time.Time `gorm:"not null"`
	// Embedding is optional until the generation pipeline fills it from [IntelKind.Content].
	// Kind is the source-family discriminator; [NewFromIntelKind] sets it from [IntelKind.Kind].
	Embedding *pgvector.Vector `gorm:"type:vector(1024)"`
	// Kind discriminates the source family (e.g. future "reddit", "website"); free-form string for now.
	Kind string `gorm:"not null;default:'';index"`
	// KindID is the source row id for that family (e.g. [github.com/lariv-in/lago/plugins/p_seer_reddit.RedditPost] ID when Kind is "reddit"); set from [IntelKind.IntelID].
	KindID uint `gorm:"not null;default:0;index"`
}

func init() {
	lago.OnDBInit("p_seer_intel.models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[Intel](db)
		return db
	})
}
