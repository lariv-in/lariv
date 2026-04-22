package p_seer_aisstream

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/lariv-in/lago/syncmap"
)

// aisMessageTypes to subscribe; must match ais-message-models.
var aisMessageTypes = []string{
	"PositionReport",
	"StandardClassBPositionReport",
	"ExtendedClassBPositionReport",
	"ShipStaticData",
}

type aisEnvelope struct {
	MessageType string          `json:"MessageType"`
	Metadata    json.RawMessage `json:"Metadata"`
	Message     json.RawMessage `json:"Message"`
}

type positionCore struct {
	UserID      float64 `json:"UserID"`
	Valid       bool    `json:"Valid"`
	Latitude    float64 `json:"Latitude"`
	Longitude   float64 `json:"Longitude"`
	Cog         float64 `json:"Cog"`
	Sog         float64 `json:"Sog"`
	TrueHeading float64 `json:"TrueHeading"`
}

type shipStaticData struct {
	UserID float64 `json:"UserID"`
	Valid  bool    `json:"Valid"`
	Name   string  `json:"Name"`
}

type extendedClassB struct {
	positionCore
	Name string `json:"Name"`
}

const staleVesselTTL = 30 * time.Minute

// vesselState is a normalized snapshot for API + map.
type vesselState struct {
	MMSI   string  `json:"mmsi"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Cog    float64 `json:"cog"`
	Sog    float64 `json:"sog"`
	Name   string  `json:"name,omitempty"`
	UpdtMs int64   `json:"-"`
}

var vesselByMMSI = &syncmap.SyncMap[string, vesselState]{}

func mmsiKey(userID float64) string {
	return fmt.Sprintf("%.0f", userID)
}

// headingForSymbol degrees 0-359; AIS may use 511 for unavailable true heading.
func headingForSymbol(cog, trueH float64) float64 {
	if trueH >= 0 && trueH < 360 {
		return trueH
	}
	if cog >= 0 && cog < 360 {
		return cog
	}
	return 0
}

func storePosition(mmsi string, lat, lng, cog, sog, trueH float64) {
	now := time.Now().UnixMilli()
	heading := headingForSymbol(cog, trueH)
	if lat == 0 && lng == 0 {
		return
	}
	prev, has := vesselByMMSI.Load(mmsi)
	v := vesselState{
		MMSI:   mmsi,
		Lat:    lat,
		Lng:    lng,
		Cog:    heading,
		Sog:    sog,
		UpdtMs: now,
	}
	if has {
		v.Name = prev.Name
	}
	vesselByMMSI.Store(mmsi, v)
}

func mergeName(mmsi, name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	now := time.Now().UnixMilli()
	prev, has := vesselByMMSI.Load(mmsi)
	if has {
		prev.Name = name
		prev.UpdtMs = now
		vesselByMMSI.Store(mmsi, prev)
		return
	}
	vesselByMMSI.Store(mmsi, vesselState{MMSI: mmsi, Name: name, UpdtMs: now})
}

// applyAISMessage parses one WebSocket JSON line into vessel updates.
func applyAISMessage(data []byte) {
	var env aisEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return
	}
	if env.MessageType == "" || len(env.Message) == 0 {
		return
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(env.Message, &body); err != nil {
		return
	}
	for key, raw := range body {
		switch key {
		case "PositionReport":
			var p positionCore
			if err := json.Unmarshal(raw, &p); err != nil || !p.Valid {
				continue
			}
			storePosition(mmsiKey(p.UserID), p.Latitude, p.Longitude, p.Cog, p.Sog, p.TrueHeading)
		case "StandardClassBPositionReport":
			var p positionCore
			if err := json.Unmarshal(raw, &p); err != nil || !p.Valid {
				continue
			}
			storePosition(mmsiKey(p.UserID), p.Latitude, p.Longitude, p.Cog, p.Sog, p.TrueHeading)
		case "ExtendedClassBPositionReport":
			var e extendedClassB
			if err := json.Unmarshal(raw, &e); err != nil || !e.Valid {
				continue
			}
			storePosition(mmsiKey(e.UserID), e.Latitude, e.Longitude, e.Cog, e.Sog, e.TrueHeading)
			if n := strings.TrimSpace(e.Name); n != "" {
				mergeName(mmsiKey(e.UserID), n)
			}
		case "ShipStaticData":
			var s shipStaticData
			if err := json.Unmarshal(raw, &s); err != nil || !s.Valid {
				continue
			}
			mergeName(mmsiKey(s.UserID), s.Name)
		default:
			_ = key
		}
	}
}

func pruneStaleVessels() {
	cut := time.Now().Add(-staleVesselTTL).UnixMilli()
	vesselByMMSI.Range(func(key string, v vesselState) bool {
		if v.UpdtMs == 0 || v.UpdtMs < cut {
			vesselByMMSI.Delete(key)
		}
		return true
	})
}

// vesselsInBbox returns vessels inside the axis-aligned box (lamin, lomin) to (lamax, lomax).
// Longitude does not handle antimeridian crossing.
func vesselsInBbox(lamin, lomin, lamax, lomax float64) []Vessel {
	pruneStaleVessels()
	lo := math.Min(lomin, lomax)
	hi := math.Max(lomin, lomax)
	la0 := math.Min(lamin, lamax)
	la1 := math.Max(lamin, lamax)
	out := make([]Vessel, 0, 64)
	cut := time.Now().Add(-staleVesselTTL).UnixMilli()
	vesselByMMSI.Range(func(_ string, v vesselState) bool {
		if v.UpdtMs < cut {
			return true
		}
		if v.Lat == 0 && v.Lng == 0 {
			return true
		}
		if v.Lat < la0 || v.Lat > la1 {
			return true
		}
		if v.Lng < lo || v.Lng > hi {
			return true
		}
		out = append(out, Vessel{MMSI: v.MMSI, Lat: v.Lat, Lng: v.Lng, Cog: v.Cog, Sog: v.Sog, Name: v.Name})
		return true
	})
	return out
}

// Vessel is the public API DTO.
type Vessel struct {
	MMSI string  `json:"mmsi"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
	Cog  float64 `json:"cog"`
	Sog  float64 `json:"sog"`
	Name string  `json:"name,omitempty"`
}
