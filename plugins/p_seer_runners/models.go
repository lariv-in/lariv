package p_seer_runners

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// Runner defines a polling interval ([Duration]) and which runnable implementation to use ([Kind]).
// Which source uses this runner is modeled on the source side (foreign key from source → runner), not here.
type Runner struct {
	gorm.Model

	Duration time.Duration `gorm:"not null"`
	Kind     string        `gorm:"not null;default:'';index"`
}

func (Runner) TableName() string {
	return "seer_runners"
}

func init() {
	lago.OnDBInit("p_seer_runners.models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[Runner](db)
		return db
	})
}
