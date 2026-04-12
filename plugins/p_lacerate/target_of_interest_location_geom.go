package p_lacerate

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"gorm.io/gorm"
)

// GeomPoint4326 scans PostGIS geometry(Point,4326) from pgx: []byte EWKB or hex text (common text format).
type GeomPoint4326 struct {
	P *ewkb.Point
}

func geometryBytesFromDriver(src interface{}) ([]byte, error) {
	switch v := src.(type) {
	case []byte:
		return v, nil
	case string:
		s := strings.TrimPrefix(strings.TrimSpace(v), `\x`)
		if decoded, err := hex.DecodeString(s); err == nil && len(decoded) > 0 {
			return decoded, nil
		}
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("geometry column: want []byte or string, got %T", src)
	}
}

// Scan implements [sql.Scanner].
func (g *GeomPoint4326) Scan(src interface{}) error {
	if g == nil {
		return fmt.Errorf("GeomPoint4326.Scan: nil receiver")
	}
	if src == nil {
		g.P = nil
		return nil
	}
	raw, err := geometryBytesFromDriver(src)
	if err != nil {
		return err
	}
	if g.P == nil {
		g.P = new(ewkb.Point)
	}
	return g.P.Scan(raw)
}

// Value implements [driver.Valuer] for rare GORM writes; prefer insertTargetOfInterestLocation for inserts.
func (g GeomPoint4326) Value() (driver.Value, error) {
	if g.P == nil {
		return nil, nil
	}
	return g.P.Value()
}

// geomPoint4326EWKB returns EWKB for a 2D point in SRID 4326 (lon, lat).
func geomPoint4326EWKB(lon, lat float64) ([]byte, error) {
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return nil, fmt.Errorf("coordinates out of WGS84 range")
	}
	gp := geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{lon, lat}).SetSRID(4326)
	var buf bytes.Buffer
	if err := ewkb.Write(&buf, ewkb.NDR, gp); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// insertTargetOfInterestLocation inserts a row with point = ST_GeomFromEWKB(decode(?, 'hex')).
// GORM expands []byte into one placeholder per byte in Raw SQL; binding hex text keeps a single argument.
func insertTargetOfInterestLocation(ctx context.Context, db *gorm.DB, targetID, intelID uint, dt time.Time, address string, pointEWKB []byte) (TargetOfInterestLocation, error) {
	var row struct {
		ID        uint
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	now := time.Now()
	hexEWKB := hex.EncodeToString(pointEWKB)
	err := db.WithContext(ctx).Raw(`
INSERT INTO targets_of_interest_locations (created_at, updated_at, deleted_at, target_of_interest_id, intel_id, datetime, address, point)
VALUES (?, ?, NULL, ?, ?, ?, ?, ST_GeomFromEWKB(decode(?, 'hex')))
RETURNING id, created_at, updated_at`,
		now, now, targetID, intelID, dt, address, hexEWKB,
	).Scan(&row).Error
	if err != nil {
		return TargetOfInterestLocation{}, err
	}
	return TargetOfInterestLocation{
		Model: gorm.Model{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		TargetOfInterestID: targetID,
		IntelID:            intelID,
		Datetime:           dt,
		Address:            address,
	}, nil
}
