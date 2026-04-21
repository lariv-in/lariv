package lago

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// PGPoint wraps PostgreSQL's geometric POINT (float8 x, float8 y).
// For WGS84 coordinates use X = longitude, Y = latitude.
type PGPoint struct {
	pgtype.Point
}

// NewPGPoint builds a valid [PGPoint] from longitude and latitude.
func NewPGPoint(lng, lat float64) PGPoint {
	return PGPoint{Point: pgtype.Point{P: pgtype.Vec2{X: lng, Y: lat}, Valid: true}}
}

// Scan implements [database/sql.Scanner]. Accepts []byte in addition to types handled by [pgtype.Point].
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
