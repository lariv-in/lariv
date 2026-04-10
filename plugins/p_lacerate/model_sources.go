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

func init() {
	lago.OnDBInit("p_lacerate.source_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[Source](db)
		return db
	})
}
