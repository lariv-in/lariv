package p_lacerate

import (
	"log/slog"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// Event is a time-stamped address and map point tied to supporting [Intel].
type Event struct {
	gorm.Model
	IntelID  uint      `gorm:"not null;index"`
	Intel    Intel     `gorm:"foreignKey:IntelID;constraint:OnDelete:CASCADE"`
	Datetime time.Time `gorm:"not null"`
	Address  string    `gorm:"type:text;not null;default:''"`
	// Populated on read; writes use insertEvent (ST_GeomFromEWKB) so pgx never binds geometry directly.
	Point GeomPoint4326 `gorm:"column:point;type:geometry(Point,4326);not null"`
}

func (Event) TableName() string { return "events" }

func init() {
	lago.OnDBInit("p_lacerate.event_model", func(db *gorm.DB) *gorm.DB {
		if db.Name() != "postgres" {
			return db
		}
		lago.RegisterModel[Event](db)
		ensureEventPointGiST(db)
		return db
	})
}

func ensureEventPointGiST(db *gorm.DB) {
	const q = `CREATE INDEX IF NOT EXISTS idx_events_point ON events USING GIST (point)`
	if err := db.Exec(q).Error; err != nil {
		slog.Error("lacerate: create GiST index on events.point", "error", err)
	}
}
