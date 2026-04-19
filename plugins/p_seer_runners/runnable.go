package p_seer_runners

import (
	"context"

	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// Runnable is implemented by types that perform a single fetch/ingest pass (job step, source pull, etc.).
// Load by name from [RegistryRunnable] for orchestration.
type Runnable interface {
	Run(ctx context.Context, db *gorm.DB) error
}

// RegistryRunnable maps stable names to runnable implementations.
var RegistryRunnable = registry.NewRegistry[Runnable]()
