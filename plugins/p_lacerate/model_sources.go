package p_lacerate

import (
	"context"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

type SourceInterface interface {
	Fetch(context.Context, *gorm.DB) ([]Intel, error)
}

type SourceDesc struct {
	Name  string
	Model SourceInterface
}

var SourceKindMap = map[string]SourceDesc{}

// RegistrySourceKind holds a constructor per [Source.Kind] that returns a new row value (e.g. &RedditSource{})
// for GORM to scan into; the dynamic type must implement [SourceInterface].
var RegistrySourceKind = registry.NewRegistry[func() SourceInterface]()

type Source struct {
	gorm.Model
	Name     string
	Kind     string
	Duration time.Duration
}

// AfterSave schedules a non-blocking fetch worker restart ([views] create/update flows use a transaction, like [Lookup.AfterSave]).
func (s *Source) AfterSave(tx *gorm.DB) error {
	if tx.Statement.SkipHooks || s == nil || s.ID == 0 {
		return nil
	}
	scheduleRestartSourceWorker(tx, s.ID)
	return nil
}

// AfterDelete stops the background fetch worker for this source row.
func (s *Source) AfterDelete(tx *gorm.DB) error {
	if s != nil && s.ID != 0 {
		StopSourceWorker(s.ID)
	}
	return nil
}

func init() {
	lago.OnDBInit("p_lacerate.source_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[Source](db)
		return db
	})
}
