package p_seer_opensky

import (
	"encoding/json"
	"fmt"
	"math"
)

// StatesEnvelope is the top-level OpenSky "all states" JSON response.
type StatesEnvelope struct {
	Time   int64         `json:"time"`
	States []StateVector `json:"states"`
}

// UnmarshalJSON treats null "states" as an empty list.
func (e *StatesEnvelope) UnmarshalJSON(b []byte) error {
	aux := &struct {
		Time   *int64           `json:"time"`
		States json.RawMessage `json:"states"`
	}{}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	if aux.Time == nil {
		return fmt.Errorf("p_seer_opensky: missing time")
	}
	e.Time = *aux.Time
	if len(aux.States) == 0 || string(aux.States) == "null" {
		e.States = nil
		return nil
	}
	var rawRows []json.RawMessage
	if err := json.Unmarshal(aux.States, &rawRows); err != nil {
		return err
	}
	e.States = make([]StateVector, 0, len(rawRows))
	for i := range rawRows {
		var row StateVector
		if err := json.Unmarshal(rawRows[i], &row); err != nil {
			return fmt.Errorf("p_seer_opensky: state row %d: %w", i, err)
		}
		e.States = append(e.States, row)
	}
	return nil
}

// StateVector is one OpenSky state row (array of mixed JSON types in the API).
// Field positions match https://openskynetwork.github.io/opensky-api/rest.html#all-state-vectors
type StateVector struct {
	Icao24         string
	Callsign       *string
	OriginCountry  *string
	TimePosition   *int64
	LastContact    *int64
	Longitude      *float64
	Latitude       *float64
	BaroAltitude   *float64
	OnGround       *bool
	Velocity       *float64
	TrueTrack      *float64
	VerticalRate   *float64
	SensorIDs      []int
	GeoAltitude    *float64
	Squawk         *string
	SPI            *bool
	PositionSource *int
	Category       *int
}

// UnmarshalJSON decodes a JSON array into [StateVector] field-by-index.
func (s *StateVector) UnmarshalJSON(b []byte) error {
	var arr []any
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	if len(arr) < 5 {
		return fmt.Errorf("p_seer_opensky: state row too short (%d)", len(arr))
	}

	icao, ok := stringish(arr[0])
	if !ok {
		return fmt.Errorf("p_seer_opensky: icao24 missing or invalid")
	}
	s.Icao24 = icao

	if arr[1] != nil {
		if cs, ok := stringish(arr[1]); ok {
			s.Callsign = &cs
		}
	}
	if arr[2] != nil {
		if oc, ok := stringish(arr[2]); ok {
			s.OriginCountry = &oc
		}
	}
	if arr[3] != nil {
		if t, ok := toInt64(arr[3]); ok {
			s.TimePosition = t
		}
	}
	if arr[4] != nil {
		if t, ok := toInt64(arr[4]); ok {
			s.LastContact = t
		}
	}
	if s.LastContact == nil {
		return fmt.Errorf("p_seer_opensky: last_contact required")
	}

	if len(arr) > 5 && arr[5] != nil {
		if v, ok := toFloat64(arr[5]); ok {
			if !math.IsNaN(v) {
				s.Longitude = &v
			}
		}
	}
	if len(arr) > 6 && arr[6] != nil {
		if v, ok := toFloat64(arr[6]); ok {
			if !math.IsNaN(v) {
				s.Latitude = &v
			}
		}
	}
	if len(arr) > 7 && arr[7] != nil {
		if v, ok := toFloat64(arr[7]); ok {
			s.BaroAltitude = &v
		}
	}
	if len(arr) > 8 && arr[8] != nil {
		if b, ok := toBool(arr[8]); ok {
			s.OnGround = &b
		}
	}
	if len(arr) > 9 && arr[9] != nil {
		if v, ok := toFloat64(arr[9]); ok {
			s.Velocity = &v
		}
	}
	if len(arr) > 10 && arr[10] != nil {
		if v, ok := toFloat64(arr[10]); ok {
			s.TrueTrack = &v
		}
	}
	if len(arr) > 11 && arr[11] != nil {
		if v, ok := toFloat64(arr[11]); ok {
			s.VerticalRate = &v
		}
	}
	if len(arr) > 12 && arr[12] != nil {
		if ids, ok := toIntSlice(arr[12]); ok {
			s.SensorIDs = ids
		}
	}
	if len(arr) > 13 && arr[13] != nil {
		if v, ok := toFloat64(arr[13]); ok {
			s.GeoAltitude = &v
		}
	}
	if len(arr) > 14 && arr[14] != nil {
		if sq, ok := stringish(arr[14]); ok {
			s.Squawk = &sq
		}
	}
	if len(arr) > 15 && arr[15] != nil {
		if b, ok := toBool(arr[15]); ok {
			s.SPI = &b
		}
	}
	if len(arr) > 16 && arr[16] != nil {
		if p, ok := toInt(arr[16]); ok {
			s.PositionSource = &p
		}
	}
	if len(arr) > 17 && arr[17] != nil {
		if c, ok := toInt(arr[17]); ok {
			s.Category = &c
		}
	}
	return nil
}

func stringish(v any) (string, bool) {
	if v == nil {
		return "", false
	}
	switch t := v.(type) {
	case string:
		return t, true
	case json.Number:
		return t.String(), true
	case float64:
		if math.Mod(t, 1) == 0 {
			return fmt.Sprintf("%.0f", t), true
		}
		return fmt.Sprintf("%g", t), true
	default:
		return fmt.Sprint(t), true
	}
}

func toInt64(v any) (*int64, bool) {
	if v == nil {
		return nil, false
	}
	switch t := v.(type) {
	case float64:
		if math.Trunc(t) == t {
			i := int64(t)
			return &i, true
		}
		return nil, false
	case json.Number:
		i, err := t.Int64()
		if err != nil {
			return nil, false
		}
		return &i, true
	case int:
		i := int64(t)
		return &i, true
	case int64:
		return &t, true
	default:
		return nil, false
	}
}

func toInt(v any) (int, bool) {
	if v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		if float64(int(t)) == t {
			return int(t), true
		}
		return 0, false
	case json.Number:
		i, err := t.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	case int:
		return t, true
	case int64:
		return int(t), true
	default:
		return 0, false
	}
}

func toFloat64(v any) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return t, true
	case json.Number:
		f, err := t.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

func toBool(v any) (bool, bool) {
	if v == nil {
		return false, false
	}
	switch t := v.(type) {
	case bool:
		return t, true
	case float64:
		return t != 0, true
	default:
		return false, false
	}
}

func toIntSlice(v any) ([]int, bool) {
	if v == nil {
		return nil, false
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]int, 0, len(arr))
	for _, e := range arr {
		if e == nil {
			continue
		}
		n, ok := toInt(e)
		if !ok {
			return nil, false
		}
		out = append(out, n)
	}
	return out, true
}
