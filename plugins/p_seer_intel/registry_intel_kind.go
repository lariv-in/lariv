package p_seer_intel

import (
	"context"
	"fmt"

	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// IntelKindLoader loads one [IntelKind] by source primary key for a single family.
// [RegistryIntelKind] keys must match [IntelKind.Kind] for rows that loader returns.
type IntelKindLoader func(ctx context.Context, db *gorm.DB, id uint) (IntelKind, error)

// RegistryIntelKind maps [IntelKind.Kind] string (e.g. "reddit") to a DB-backed loader.
var RegistryIntelKind = registry.NewRegistry[IntelKindLoader]()

// LoadIntelKind resolves [IntelKind] using [RegistryIntelKind].
func LoadIntelKind(ctx context.Context, db *gorm.DB, kind string, id uint) (IntelKind, error) {
	if kind == "" {
		return nil, fmt.Errorf("p_seer_intel: LoadIntelKind: empty kind")
	}
	loader, ok := RegistryIntelKind.Get(kind)
	if !ok {
		return nil, fmt.Errorf("p_seer_intel: LoadIntelKind: unknown kind %q", kind)
	}
	k, err := loader(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if k.Kind() != kind {
		return nil, fmt.Errorf("p_seer_intel: LoadIntelKind: kind mismatch: registry key %q, instance %q", kind, k.Kind())
	}
	return k, nil
}
