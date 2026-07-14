package fields

import (
	"database/sql/driver"
	"encoding"
	"fmt"
	"math/big"
	"strings"
)

// DecimalSix represents a decimal numeric amount rounded to exactly 6 decimal places.
// It is intended for high-precision monetary-style fields or coordinate points.
// It maps to database NUMERIC storage columns and uses Go's standard [*big.Rat] representation.
type DecimalSix struct {
	// R represents the underlying high-precision big rational number value.
	R *big.Rat
}

var (
	_ encoding.TextMarshaler   = DecimalSix{}
	_ encoding.TextUnmarshaler = (*DecimalSix)(nil)
	_ driver.Valuer            = DecimalSix{}
)

// NormalizeDecimals yields a new DecimalSix structure with its big rational value rounded to exactly 6 decimal places.
func (p DecimalSix) NormalizeDecimals() DecimalSix {
	r := new(big.Rat)
	if p.R == nil {
		r = big.NewRat(0, 1)
	} else {
		r = r.Set(p.R)
	}
	r = r.Mul(r, big.NewRat(1000000, 1))
	r.SetInt(new(big.Int).Div(r.Num(), r.Denom()))
	r = r.Quo(r, big.NewRat(1000000, 1))
	return DecimalSix{R: r}
}

// MarshalText serializes the rational number into a byte slice formatted with 6 decimal places.
func (p DecimalSix) MarshalText() ([]byte, error) {
	r := p.NormalizeDecimals().R
	return []byte(r.FloatString(6)), nil
}

// UnmarshalText deserializes the byte slice formatted string into the high-precision rational number.
func (p *DecimalSix) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		p.R = big.NewRat(0, 1)
		return nil
	}
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return fmt.Errorf("invalid points value %q", s)
	}
	*p = DecimalSix{R: r}.NormalizeDecimals()
	return nil
}

// Value implements the database driver Valuer interface to save decimal numbers as numeric database strings.
func (p DecimalSix) Value() (driver.Value, error) {
	return p.NormalizeDecimals().R.FloatString(6), nil
}

// Scan implements the sql Scanner interface to populate decimal structures from database columns.
func (p *DecimalSix) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		p.R = big.NewRat(0, 1)
		return nil
	case []byte:
		return p.UnmarshalText(v)
	case string:
		return p.UnmarshalText([]byte(v))
	case int64:
		p.R = big.NewRat(v, 1)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into PointsDecimal", src)
	}
}

// String returns a formatted fixed 6-decimal string representation suitable for UI views.
func (p DecimalSix) String() string {
	b, err := p.MarshalText()
	if err != nil {
		return "0.000000"
	}
	return string(b)
}
