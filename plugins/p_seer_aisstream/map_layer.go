package p_seer_aisstream

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const aisStreamMapVesselsKey = "seer_aisstream.map_vessels"
const aisStreamMapMaxVessels = 20000

type aisStreamMapVessel struct {
	ID         uint    `json:"id"`
	MMSI       string  `json:"mmsi"`
	Title      string  `json:"title"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	Heading    float64 `json:"heading"`
	SOG        float64 `json:"sog"`
	TimeUTC    int64   `json:"timeUtc"`
	DetailPath string  `json:"detailPath"`
}

type aisStreamMapLayer struct{}

func (aisStreamMapLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		db, err := getters.DBFromContext(ctx)
		if err != nil {
			slog.Error("p_seer_aisstream: map layer: db from context", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": err})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		vessels, err := buildAISStreamMapVessels(ctx, db)
		if err != nil {
			slog.Error("p_seer_aisstream: map layer: load", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{"_global": fmt.Errorf("map vessels: %w", err)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, aisStreamMapVesselsKey, vessels)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func buildAISStreamMapVessels(ctx context.Context, db *gorm.DB) ([]aisStreamMapVessel, error) {
	if db == nil {
		return nil, nil
	}
	cutoff := time.Time{}
	if w := Config.MapLastContactWindow(); w > 0 {
		cutoff = time.Now().Add(-w)
	}
	var rows []AISStreamMessage
	q := db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("position IS NOT NULL").
		Order("received_at DESC").
		Limit(aisStreamMapMaxVessels)
	if !cutoff.IsZero() {
		q = q.Where("received_at >= ?", cutoff)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	out := make([]aisStreamMapVessel, 0, len(rows))
	for i := range rows {
		row := rows[i]
		if !row.Position.Valid {
			continue
		}
		key := strings.TrimSpace(row.MMSI)
		if key == "" {
			key = strconv.FormatUint(uint64(row.ID), 10)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		title := strings.TrimSpace(row.ShipName)
		if title == "" {
			title = strings.TrimSpace(row.MMSI)
		}
		if title == "" {
			title = row.MessageType
		}
		heading := 0.0
		if row.Heading != nil {
			heading = *row.Heading
		}
		sog := 0.0
		if row.SOG != nil {
			sog = *row.SOG
		}
		timeUTC := int64(0)
		if row.TimeUTC != nil {
			timeUTC = row.TimeUTC.Unix()
		}
		out = append(out, aisStreamMapVessel{
			ID:         row.ID,
			MMSI:       row.MMSI,
			Title:      title,
			Lat:        row.Position.P.Y,
			Lng:        row.Position.P.X,
			Heading:    heading,
			SOG:        sog,
			TimeUTC:    timeUTC,
			DetailPath: aisStreamMessageDetailPath(row.ID),
		})
	}
	return out, nil
}

func aisStreamMessageDetailPath(id uint) string {
	if id == 0 {
		return ""
	}
	return AppUrl + "messages/" + strconv.FormatUint(uint64(id), 10) + "/"
}
