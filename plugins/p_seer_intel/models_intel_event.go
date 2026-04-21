package p_seer_intel

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// IntelEvent holds location and time extracted from an [Intel] summary (one row per intel by default).
type IntelEvent struct {
	gorm.Model

	IntelID uint   `gorm:"not null;uniqueIndex"`
	Intel   *Intel `gorm:"foreignKey:IntelID"`

	// Address is a free-form place string suitable for the Google Geocoding API.
	Address string `gorm:"type:text;not null;default:''"`
	// Datetime is the inferred event time from the summary (not ingest time).
	Datetime time.Time `gorm:"not null"`
	// Location is WGS84 as PostgreSQL point: X = longitude, Y = latitude.
	Location lago.PGPoint `gorm:"type:point"`
}

func (e *IntelEvent) BeforeCreate(tx *gorm.DB) error {
	addr := strings.TrimSpace(e.Address)
	if addr == "" {
		return nil
	}
	key := strings.TrimSpace(IntelGenAI.GeocodingAPIKey)
	if key == "" {
		return nil
	}
	ctx := tx.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}
	lat, lng, err := geocodeGoogleMaps(ctx, key, addr)
	if err != nil {
		slog.Warn("p_seer_intel: geocode intel event", "error", err, "address", addr)
		return nil
	}
	e.Location = lago.NewPGPoint(lng, lat)
	return nil
}
