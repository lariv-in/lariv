package lariv

import (
	"database/sql/driver"
	"fmt"
	"net/url"
)

// PageURL represents a GORM-compatible database-persisted URL wrapper.
// It embeds standard Go [url.URL] allowing callers to directly access standard URL fields and methods (e.g. [url.URL.String]).
//
// Use Cases:
//   - Storing web resources urls (e.g. avatar images, callback webhooks, external client logins) in database models.
//
// Example:
//
//	type ClientApp struct {
//		gorm.Model
//		Homepage lariv.PageURL
//	}
type PageURL struct {
	// URL embeds standard url.URL structures.
	url.URL
}

// URLPtr returns a pointer to a copy of the embedded URL (or nil if empty / invalid for use as *url.URL).
// URLPtr yields a standard url.URL pointer to a copy of the embedded URL structure.
// Returns nil if the URL Host string is empty or invalid.
func (p *PageURL) URLPtr() *url.URL {
	if p == nil || p.Host == "" {
		return nil
	}
	return &p.URL
}

// SetFromURL overrides the embedded URL properties using fields from the provided *url.URL object.
// Providing nil clears the url properties to zero.
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

// Value implements the database driver Valuer interface to save URL properties as database text strings.
func (p PageURL) Value() (driver.Value, error) {
	if p.Host == "" {
		return "", nil
	}
	return p.String(), nil
}

// Scan implements the SQL Scanner interface to populate URL properties from database columns.
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
		return fmt.Errorf("lariv: PageURL Scan: unexpected %T", value)
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

// UnmarshalText implements encoding.TextUnmarshaler to parse TOML configurations and form bindings.
func (p *PageURL) UnmarshalText(text []byte) error {
	return p.Scan(text)
}

// MarshalText implements encoding.TextMarshaler to serialize URL objects to byte strings.
func (p PageURL) MarshalText() ([]byte, error) {
	v, err := p.Value()
	if err != nil {
		return nil, err
	}
	s, _ := v.(string)
	return []byte(s), nil
}
