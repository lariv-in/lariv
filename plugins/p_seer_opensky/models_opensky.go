package p_seer_opensky

import (
	"encoding/json"

	"github.com/lariv-in/lago/lago"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OpenSkyStatesTable is the GORM/Postgres table for aircraft state rows.
const OpenSkyStatesTable = "seer_opensky_states"

// OpenSkyState stores one state vector (API row) with deduplication on Icao24+LastContact.
type OpenSkyState struct {
	gorm.Model

	SnapshotTime int64 `gorm:"not null;index"`

	Icao24         string  `gorm:"size:8;not null;uniqueIndex:ux_opensky_icao24_last_cont;index:idx_opensky_map_icao_last,priority:1,where:deleted_at IS NULL"`
	LastContact    int64   `gorm:"not null;uniqueIndex:ux_opensky_icao24_last_cont;index:idx_opensky_map_icao_last,priority:2,sort:desc,where:deleted_at IS NULL;index:idx_opensky_map_last_contact,sort:desc,where:deleted_at IS NULL"`
	Callsign       *string `gorm:"size:8"`
	OriginCountry  *string `gorm:"size:64"`
	TimePosition *int64
	// Position is WGS84 (lng, lat); persisted as PostgreSQL point. Use [Longitude]/[Latitude] for forms and API mapping.
	Position  lago.PGPoint `gorm:"type:point"`
	Longitude float64     `gorm:"-"` // form roundtrip; see [OpenSkyState.syncPositionFromPoint] / [BeforeSave]
	Latitude  float64     `gorm:"-"`
	BaroAltitude   *float64
	OnGround       *bool
	Velocity       *float64
	TrueTrack      *float64
	VerticalRate   *float64
	// Sensors holds JSON array of receiver id ints (index 12 in the API).
	Sensors     datatypes.JSON `gorm:"type:json"`
	SensorsText string         `gorm:"-"` // form round-trip JSON array text; not persisted
	GeoAltitude    *float64
	Squawk         *string `gorm:"size:8"`
	SPI            *bool
	PositionSource *int
	Category       *int
}

func (OpenSkyState) TableName() string {
	return OpenSkyStatesTable
}

func (m *OpenSkyState) AfterFind(_ *gorm.DB) error {
	if m == nil {
		return nil
	}
	if len(m.Sensors) == 0 {
		m.SensorsText = "[]"
	} else {
		m.SensorsText = string(m.Sensors)
	}
	m.syncPositionFromPoint()
	return nil
}

// BeforeSave encodes [SensorsText] into [Sensors] and [Longitude]/[Latitude] into [Position].
func (m *OpenSkyState) BeforeSave(_ *gorm.DB) error {
	if m == nil {
		return nil
	}
	if m.SensorsText != "" {
		var arr []int
		if err := json.Unmarshal([]byte(m.SensorsText), &arr); err != nil {
			return err
		}
		b, err := json.Marshal(arr)
		if err != nil {
			return err
		}
		m.Sensors = datatypes.JSON(b)
	}
	if openskyValidLatLng(m.Latitude, m.Longitude) {
		m.Position = lago.NewPGPoint(m.Longitude, m.Latitude)
	} else {
		m.Position = lago.PGPoint{}
	}
	return nil
}

func (m *OpenSkyState) syncPositionFromPoint() {
	if m.Position.Valid {
		m.Longitude = m.Position.P.X
		m.Latitude = m.Position.P.Y
	} else {
		m.Longitude = 0
		m.Latitude = 0
	}
}

// openskyValidLatLng is false for (0,0) and coordinates outside WGS84 (same idea as p_seer_gdelt [geo.go]).
func openskyValidLatLng(lat, lng float64) bool {
	if lat == 0 && lng == 0 {
		return false
	}
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return false
	}
	return true
}
