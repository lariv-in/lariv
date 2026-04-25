package lago

import (
	"database/sql/driver"
	"fmt"
	"net/url"
)

// PageURL is a persisted HTTP(S) URL (GORM text + mapstructure / encoding.Text from strings).
// It embeds [url.URL] so callers use normal URL methods ([url.URL.String], [url.URL.Hostname], etc.).
type PageURL struct {
	url.URL
}

// URLPtr returns a pointer to a copy of the embedded URL (or nil if empty / invalid for use as *url.URL).
func (p *PageURL) URLPtr() *url.URL {
	if p == nil || p.Host == "" {
		return nil
	}
	return new(p.URL)
}

// SetFromURL copies u into p (nil clears to zero).
func (p *PageURL) SetFromURL(u *url.URL) {
	if p == nil {
		return
	}
	if u == nil {
		*p = PageURL{}
		return
	}
	p.URL = *u
}

// Value implements [driver.Valuer] for GORM/SQL.
func (p PageURL) Value() (driver.Value, error) {
	if p.Host == "" {
		return "", nil
	}
	return p.String(), nil
}

// Scan implements [sql.Scanner] for GORM/SQL.
func (p *PageURL) Scan(value any) error {
	if value == nil {
		*p = PageURL{}
		return nil
	}
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("lago: PageURL Scan: unexpected %T", value)
	}
	if s == "" {
		*p = PageURL{}
		return nil
	}
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	if u == nil {
		*p = PageURL{}
		return nil
	}
	p.URL = *u
	return nil
}

// UnmarshalText implements [encoding.TextUnmarshaler] for TOML/mapstructure form decode.
func (p *PageURL) UnmarshalText(text []byte) error {
	return p.Scan(text)
}

// MarshalText implements [encoding.TextMarshaler].
func (p PageURL) MarshalText() ([]byte, error) {
	v, err := p.Value()
	if err != nil {
		return nil, err
	}
	s, _ := v.(string)
	return []byte(s), nil
}
