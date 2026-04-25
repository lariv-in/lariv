package p_seer_opensky

import (
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
)

// ToOpenSkyState maps an API [StateVector] and response envelope [time] to a new [OpenSkyState].
// Returns nil if the row should be skipped.
func ToOpenSkyState(sv *StateVector, snapshotTime int64) (*OpenSkyState, error) {
	if sv == nil || sv.Icao24 == "" || sv.LastContact == nil {
		return nil, nil
	}
	m := &OpenSkyState{
		SnapshotTime:   snapshotTime,
		Icao24:         sv.Icao24,
		LastContact:    *sv.LastContact,
		Callsign:       sv.Callsign,
		OriginCountry:  sv.OriginCountry,
		TimePosition:   sv.TimePosition,
		BaroAltitude:   sv.BaroAltitude,
		OnGround:       sv.OnGround,
		Velocity:       sv.Velocity,
		TrueTrack:      sv.TrueTrack,
		VerticalRate:   sv.VerticalRate,
		GeoAltitude:    sv.GeoAltitude,
		Squawk:         sv.Squawk,
		SPI:            sv.SPI,
		PositionSource: sv.PositionSource,
		Category:       sv.Category,
	}
	if len(sv.SensorIDs) > 0 {
		b, err := json.Marshal(sv.SensorIDs)
		if err != nil {
			return nil, fmt.Errorf("sensors: %w", err)
		}
		m.Sensors = datatypes.JSON(b)
	}
	if sv.Longitude != nil && sv.Latitude != nil {
		m.Longitude, m.Latitude = *sv.Longitude, *sv.Latitude
	}
	// [BeforeSave] persists [Position] from Longitude/Latitude.
	return m, nil
}
