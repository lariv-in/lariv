package p_seer_runners

import (
	"context"

	"gorm.io/gorm"
)

// Runnable is implemented by types that perform a single fetch/ingest pass. A [Runner] will invoke Run
// on a schedule; sources reference runners from their own models; dispatch by [Runner.Kind] can be wired later.
type Runnable interface {
	Run(ctx context.Context, db *gorm.DB) error
}
