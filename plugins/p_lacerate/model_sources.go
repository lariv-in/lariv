package p_lacerate

import (
	"context"
	"time"

	"github.com/lariv-in/lago/lago"
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
