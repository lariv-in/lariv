package p_seer_opensky

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// Context key for [openSkyMapAircraft] JSON (map page script).
const openskyMapAircraftKey = "seer_opensky.map_aircraft"

// Max distinct aircraft on the map (by latest last_contact). Fetches up to 2 state rows per aircraft.
const openskyMapMaxAircraft = 5000

// openSkyMapAircraft is the JSON payload for the map script (one feature per icao24).
type openSkyMapAircraft struct {
	Icao24      string  `json:"icao24"`
	ID          uint    `json:"id"` // GORM id of the latest (newest last_contact) row
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Heading     float64 `json:"heading"` // degrees, 0 = north, for icon-rotate
	VelocityMps float64 `json:"velocityMps"`
	Title       string  `json:"title"`
	DetailPath  string  `json:"detailPath"`
}

// openSkyMapRow is one row from the CTE (rn = 1 newest, 2 = second by last_contact).
type openSkyMapRow struct {
	ID          uint
	Icao24      string
	LastContact int64
	Lng         float64
	Lat         float64
	Velocity    *float64
	TrueTrack   *float64
	Callsign    *string
	Rn          int
}

type openSkyMapLayer struct{}

func (openSkyMapLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		db, err := getters.DBFromContext(ctx)
		if err != nil {
			slog.Error("p_seer_opensky: map layer: db from context", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		aircraft, err := buildOpenSkyMapAircraft(ctx, db)
		if err != nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
			// Client closed or deadline; avoid _global + Render panic (broken pipe) for gone clients.
			aircraft, err = []openSkyMapAircraft{}, nil
		}
		if err != nil {
			slog.Error("p_seer_opensky: map layer: load", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": fmt.Errorf("map aircraft: %w", err),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, openskyMapAircraftKey, aircraft)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// buildOpenSkyMapAircraft returns one entry per icao (latest point), ordered by that row's last_contact desc.
// PostgreSQL only; other dialects return nil, nil.
func buildOpenSkyMapAircraft(ctx context.Context, db *gorm.DB) ([]openSkyMapAircraft, error) {
	if db == nil {
		return nil, nil
	}
	if db.Dialector.Name() != "postgres" {
		slog.Info("p_seer_opensky: map aircraft skipped: requires PostgreSQL", "dialector", db.Dialector.Name())
		return nil, nil
	}
	sql := `
WITH latest AS (
  SELECT DISTINCT ON (s.icao24)
    s.icao24,
    s.last_contact,
    s.id,
    s.velocity,
    s.on_ground
  FROM seer_opensky_states s
  WHERE s.deleted_at IS NULL
    AND COALESCE(BTRIM(s.icao24), '') <> ''
    AND s."position" IS NOT NULL
    AND ((s."position")[0] <> 0 OR (s."position")[1] <> 0)
    AND s.last_contact >= ?
  ORDER BY s.icao24, s.last_contact DESC, s.id DESC
),
top_icao AS (
  SELECT
    l.icao24,
    l.last_contact AS last_mx
  FROM latest l
  WHERE
    l.on_ground = false
    AND l.velocity IS NOT NULL
    AND l.velocity <> 0
  ORDER BY last_mx DESC
  LIMIT ` + strconv.Itoa(openskyMapMaxAircraft) + `
),
r AS (
  SELECT
    q.id, q.icao24, q.last_contact, q.velocity, q.true_track, q.callsign,
    q.lng, q.lat, q.rn
  FROM top_icao t
  CROSS JOIN LATERAL (
    SELECT
      x.id, x.icao24, x.last_contact, x.velocity, x.true_track, x.callsign,
      x.lng, x.lat,
      ROW_NUMBER() OVER (ORDER BY x.last_contact DESC, x.id DESC) AS rn
    FROM (
      SELECT
        s.id, s.icao24, s.last_contact, s.velocity, s.true_track, s.callsign,
        (s."position")[0] AS lng, (s."position")[1] AS lat
      FROM seer_opensky_states s
      WHERE s.deleted_at IS NULL
        AND s.icao24 = t.icao24
        AND s."position" IS NOT NULL
        AND ((s."position")[0] <> 0 OR (s."position")[1] <> 0)
      ORDER BY s.last_contact DESC, s.id DESC
      LIMIT 2
    ) x
  ) q
)
SELECT id, icao24, last_contact, velocity, true_track, callsign, lng, lat, rn FROM r WHERE rn <= 2
ORDER BY icao24, last_contact ASC, id ASC`
	cutoff := int64(0)
	if w := Config.MapLastContactWindow(); w > 0 {
		cutoff = time.Now().Add(-w).Unix()
	}
	var rows []openSkyMapRow
	// Read-only; use WithoutCancel so devtools/livereload/nested requests do not
	// drop this query mid-flight (GORM "context canceled" + empty map).
	qctx := context.WithoutCancel(ctx)
	if err := db.WithContext(qctx).Raw(sql, cutoff).Scan(&rows).Error; err != nil {
		return nil, err
	}
	by := make(map[string][]openSkyMapRow, 256)
	for i := range rows {
		icao := rows[i].Icao24
		by[icao] = append(by[icao], rows[i])
	}
	type scored struct {
		air  openSkyMapAircraft
		rank int64
	}
	var scoredList []scored
	for icao, arr := range by {
		if len(arr) == 0 {
			continue
		}
		sort.Slice(arr, func(i, j int) bool {
			if arr[i].LastContact == arr[j].LastContact {
				return arr[i].ID < arr[j].ID
			}
			return arr[i].LastContact < arr[j].LastContact
		})
		newest := &arr[len(arr)-1]
		heading := 0.0
		if len(arr) >= 2 {
			older := &arr[0]
			heading = initialBearingDeg(older.Lat, older.Lng, newest.Lat, newest.Lng)
		} else if newest.TrueTrack != nil {
			heading = *newest.TrueTrack
		}
		if newest.Velocity == nil || *newest.Velocity == 0 {
			continue
		}
		vel := *newest.Velocity
		title := icao
		if newest.Callsign != nil {
			if cs := strings.TrimSpace(*newest.Callsign); cs != "" {
				title = cs + " (" + icao + ")"
			}
		}
		scoredList = append(scoredList, scored{
			air: openSkyMapAircraft{
				Icao24:      icao,
				ID:          newest.ID,
				Lat:         newest.Lat,
				Lng:         newest.Lng,
				Heading:     heading,
				VelocityMps: vel,
				Title:       title,
				DetailPath:  openSkyStateDetailPath(newest.ID),
			},
			rank: newest.LastContact,
		})
	}
	sort.Slice(scoredList, func(i, j int) bool { return scoredList[i].rank > scoredList[j].rank })
	out := make([]openSkyMapAircraft, 0, len(scoredList))
	for _, s := range scoredList {
		out = append(out, s.air)
	}
	return out, nil
}

func openSkyStateDetailPath(id uint) string {
	if id == 0 {
		return ""
	}
	return AppUrl + "states/" + strconv.FormatUint(uint64(id), 10) + "/"
}
