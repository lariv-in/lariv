package p_seer_gdelt

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// Max events loaded for the map (most recent by primary key). Keeps payload bounded.
const gdeltMapMaxEvents = 5000

// Context key for []gdeltMapMarker produced by gdeltMapLayer.
const gdeltMapMarkersKey = "seer_gdelt.map_markers"

// gdeltMapMarker is the JSON payload for the map page script (not a full Event).
type gdeltMapMarker struct {
	EventID    uint    `json:"eventId"`
	Kind       string  `json:"kind"` // actor1 | actor2 | action
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	Title      string  `json:"title"`
	DetailPath string  `json:"detailPath"` // app-relative URL, e.g. /seer-gdelt/events/1/
}

type gdeltMapLayer struct{}

func (gdeltMapLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		db, err := getters.DBFromContext(ctx)
		if err != nil {
			slog.Error("p_seer_gdelt: map layer: db from context", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		events, err := gorm.G[Event](db.WithContext(ctx)).
			Order("id DESC").
			Limit(gdeltMapMaxEvents).
			Find(ctx)
		if err != nil {
			slog.Error("p_seer_gdelt: map layer: load events", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": fmt.Errorf("failed to load events for map: %w", err),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		markers := buildGDELTMapMarkers(events)
		ctx = context.WithValue(ctx, gdeltMapMarkersKey, markers)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func buildGDELTMapMarkers(events []Event) []gdeltMapMarker {
	out := make([]gdeltMapMarker, 0, len(events)*3)
	for _, ev := range events {
		basePath := eventDetailPath(ev.ID)
		if gdeltValidLatLng(ev.Actor1GeoLat, ev.Actor1GeoLong) {
			out = append(out, gdeltMapMarker{
				EventID:    ev.ID,
				Kind:       "actor1",
				Lat:        ev.Actor1GeoLat,
				Lng:        ev.Actor1GeoLong,
				Title:      gdeltMapPopupTitle("Actor 1", ev.EventCode, ev.Actor1GeoFullName),
				DetailPath: basePath,
			})
		}
		if gdeltValidLatLng(ev.Actor2GeoLat, ev.Actor2GeoLong) {
			out = append(out, gdeltMapMarker{
				EventID:    ev.ID,
				Kind:       "actor2",
				Lat:        ev.Actor2GeoLat,
				Lng:        ev.Actor2GeoLong,
				Title:      gdeltMapPopupTitle("Actor 2", ev.EventCode, ev.Actor2GeoFullName),
				DetailPath: basePath,
			})
		}
		if ev.ActionGeoPoint.Valid && gdeltValidLatLng(ev.ActionGeoPoint.P.Y, ev.ActionGeoPoint.P.X) {
			out = append(out, gdeltMapMarker{
				EventID:    ev.ID,
				Kind:       "action",
				Lat:        ev.ActionGeoPoint.P.Y,
				Lng:        ev.ActionGeoPoint.P.X,
				Title:      gdeltMapPopupTitle("Action", ev.EventCode, ev.ActionGeoFullName),
				DetailPath: basePath,
			})
		}
	}
	return out
}

func eventDetailPath(id uint) string {
	return AppUrl + "events/" + strconv.FormatUint(uint64(id), 10) + "/"
}

func gdeltMapPopupTitle(role, eventCode, geoName string) string {
	geoName = strings.TrimSpace(geoName)
	eventCode = strings.TrimSpace(eventCode)
	switch {
	case eventCode != "" && geoName != "":
		return role + ": " + eventCode + " — " + geoName
	case eventCode != "":
		return role + ": " + eventCode
	case geoName != "":
		return role + ": " + geoName
	default:
		return role
	}
}
