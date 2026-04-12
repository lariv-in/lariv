package p_lacerate

import (
	"log/slog"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// TargetOfInterestLocation is a time-stamped address and map point for [TargetOfInterest], tied to supporting [Intel].
type TargetOfInterestLocation struct {
	gorm.Model
	TargetOfInterestID uint             `gorm:"not null;index"`
	TargetOfInterest   TargetOfInterest `gorm:"foreignKey:TargetOfInterestID;constraint:OnDelete:CASCADE"`
	IntelID            uint             `gorm:"not null;index"`
	Intel              Intel            `gorm:"foreignKey:IntelID;constraint:OnDelete:CASCADE"`
	Datetime           time.Time        `gorm:"not null"`
	Address            string           `gorm:"type:text;not null;default:''"`
	// Populated on read; writes use insertTargetOfInterestLocation (ST_GeomFromEWKB) so pgx never binds geometry directly.
	Point GeomPoint4326 `gorm:"column:point;type:geometry(Point,4326);not null"`
}

func (TargetOfInterestLocation) TableName() string { return "targets_of_interest_locations" }

func init() {
	lago.OnDBInit("p_lacerate.target_of_interest_location_model", func(db *gorm.DB) *gorm.DB {
		if db.Name() != "postgres" {
			return db
		}
		lago.RegisterModel[TargetOfInterestLocation](db)
		ensureTargetOfInterestLocationPointGiST(db)
		return db
	})
}

func ensureTargetOfInterestLocationPointGiST(db *gorm.DB) {
	const q = `CREATE INDEX IF NOT EXISTS idx_targets_of_interest_locations_point ON targets_of_interest_locations USING GIST (point)`
	if err := db.Exec(q).Error; err != nil {
		slog.Error("lacerate: create GiST index on targets_of_interest_locations.point", "error", err)
	}
}
