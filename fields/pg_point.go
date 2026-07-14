package fields

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// PGPoint wraps PostgreSQL's geometric POINT type (representing float8 X and Y coordinates).
// When storing WGS84 geographic coordinates, use X to represent longitude and Y to represent latitude.
type PGPoint struct {
	// Point embeds pgtype.Point containing Vec2 structures and validation markers.
	pgtype.Point
}

// NewPGPoint constructs a valid [PGPoint] wrapper using raw longitude and latitude floats.
func NewPGPoint(lng, lat float64) PGPoint {
	return PGPoint{Point: pgtype.Point{P: pgtype.Vec2{X: lng, Y: lat}, Valid: true}}
}

// Scan implements the database sql Scanner interface.
// It supports scanning standard pgtype objects as well as raw byte arrays.
func (p *PGPoint) Scan(src any) error {
	if p == nil {
		return nil
	}
	switch v := src.(type) {
	case nil:
		*p = PGPoint{}
		return nil
	case []byte:
		return p.Point.Scan(string(v))
	default:
		return p.Point.Scan(v)
	}
}
