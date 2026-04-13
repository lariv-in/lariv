package p_lacerate

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// mapEventMaxCosineDistance is the max pgvector cosine distance (<=>) for assigning an event's intel
// to a target of interest on the map. Same metric as [searchTargetsOfInterestByEmbedding]: 0 = identical, 2 = opposite.
const mapEventMaxCosineDistance = 0.45

const ctxKeyLacerateMapData = "lacerateMapData"

type eventMapMarker struct {
	EventID              uint    `json:"eventId"`
	Lat                  float64 `json:"lat"`
	Lng                  float64 `json:"lng"`
	Address              string  `json:"address"`
	Datetime             string  `json:"datetime"`
	Title                string  `json:"title"`
	IntelID              uint    `json:"intelId"`
	IntelPreview         string  `json:"intelPreview"`
	IntelURL             string  `json:"intelUrl"`
	TargetOfInterestName string  `json:"targetOfInterestName,omitempty"`
	TargetOfInterestURL  string  `json:"targetOfInterestUrl,omitempty"`
}

type mapLayerGroup struct {
	Key                string           `json:"key"`
	Label              string           `json:"label"`
	TargetOfInterestID uint             `json:"targetOfInterestId,omitempty"`
	TargetURL          string           `json:"targetUrl,omitempty"`
	Markers            []eventMapMarker `json:"markers"`
}

type lacerateMapData struct {
	Layers             []mapLayerGroup
	UnsupportedMessage string
}

type lacerateMapLayer struct{}

func intelEmbeddingIsEffectivelyZero(v pgvector.Vector) bool {
	s := v.Slice()
	for _, x := range s {
		if x != 0 {
			return false
		}
	}
	return true
}

func formatLacerateMapDatetime(ctx context.Context, dt time.Time) string {
	if tz, ok := ctx.Value("$tz").(*time.Location); ok && tz != nil {
		return dt.In(tz).Format(time.DateTime)
	}
	return dt.Format(time.DateTime)
}

func previewLacerateMapIntel(content string, intelID uint) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Sprintf("Intel #%d", intelID)
	}
	if len(content) > 180 {
		return content[:177] + "..."
	}
	return content
}

func routePathWithID(name string, id uint) string {
	route, ok := lago.RegistryRoute.Get(name)
	if !ok {
		return ""
	}
	return strings.ReplaceAll(route.Path, "{id}", strconv.FormatUint(uint64(id), 10))
}

func lacerateMapDataFromEvents(ctx context.Context, db *gorm.DB) lacerateMapData {
	if db == nil {
		slog.Error("lacerate: map: missing db in context")
		return lacerateMapData{UnsupportedMessage: "Database unavailable."}
	}
	if db.Name() != "postgres" {
		return lacerateMapData{UnsupportedMessage: "Map needs PostgreSQL/PostGIS-backed location data."}
	}

	var rows []Event
	if err := db.WithContext(ctx).
		Preload("Intel").
		Order("datetime DESC, id DESC").
		Find(&rows).Error; err != nil {
		slog.Error("lacerate: map: load events", "error", err)
		return lacerateMapData{UnsupportedMessage: "Could not load plotted events."}
	}

	dbq := db.WithContext(ctx)
	buckets := make(map[string][]eventMapMarker)

	for _, row := range rows {
		if row.Point.P == nil {
			slog.Error("lacerate: map: event missing geometry", "event_id", row.ID)
			continue
		}

		addr := strings.TrimSpace(row.Address)
		title := addr
		if title == "" {
			title = fmt.Sprintf("Event #%d", row.ID)
		}

		bucketKey := "uncategorized"
		var toiName, toiURL string

		intel := row.Intel
		if intel.ID != 0 && !intelEmbeddingIsEffectivelyZero(intel.Embedding) {
			toi, dist, ok, err := nearestTargetOfInterestByEmbedding(dbq, intel.Embedding)
			if err != nil {
				slog.Error("lacerate: map: nearest target for event", "error", err, "event_id", row.ID, "intel_id", intel.ID)
			} else if ok && dist <= mapEventMaxCosineDistance {
				bucketKey = fmt.Sprintf("toi-%d", toi.ID)
				toiName = strings.TrimSpace(toi.Name)
				if toiName == "" {
					toiName = fmt.Sprintf("Target #%d", toi.ID)
				}
				toiURL = routePathWithID("lacerate.TargetOfInterestDetailRoute", toi.ID)
			}
		}

		m := eventMapMarker{
			EventID:      row.ID,
			Lat:          row.Point.P.Y(),
			Lng:          row.Point.P.X(),
			Address:      addr,
			Datetime:     formatLacerateMapDatetime(ctx, row.Datetime),
			Title:        title,
			IntelID:      row.IntelID,
			IntelPreview: previewLacerateMapIntel(row.Intel.Content, row.IntelID),
			IntelURL:     routePathWithID("lacerate.IntelDetailRoute", row.IntelID),
		}
		if bucketKey != "uncategorized" {
			m.TargetOfInterestName = toiName
			m.TargetOfInterestURL = toiURL
		}
		buckets[bucketKey] = append(buckets[bucketKey], m)
	}

	var toiBucketKeys []string
	for k := range buckets {
		if strings.HasPrefix(k, "toi-") {
			toiBucketKeys = append(toiBucketKeys, k)
		}
	}
	sort.Slice(toiBucketKeys, func(i, j int) bool {
		idi, _ := strconv.ParseUint(strings.TrimPrefix(toiBucketKeys[i], "toi-"), 10, 32)
		idj, _ := strconv.ParseUint(strings.TrimPrefix(toiBucketKeys[j], "toi-"), 10, 32)
		return idi < idj
	})

	layers := make([]mapLayerGroup, 0, len(buckets))
	for _, bk := range toiBucketKeys {
		ms := buckets[bk]
		if len(ms) == 0 {
			continue
		}
		idu, _ := strconv.ParseUint(strings.TrimPrefix(bk, "toi-"), 10, 32)
		tid := uint(idu)
		label := ms[0].TargetOfInterestName
		if label == "" {
			label = fmt.Sprintf("Target #%d", tid)
		}
		label = fmt.Sprintf("%s (#%d)", label, tid)
		layers = append(layers, mapLayerGroup{
			Key:                bk,
			Label:              label,
			TargetOfInterestID: tid,
			TargetURL:          ms[0].TargetOfInterestURL,
			Markers:            ms,
		})
	}
	if unc := buckets["uncategorized"]; len(unc) > 0 {
		layers = append(layers, mapLayerGroup{
			Key:     "uncategorized",
			Label:   "Uncategorized",
			Markers: unc,
		})
	}

	return lacerateMapData{Layers: layers}
}

func (lacerateMapLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, _ := getters.DBFromContext(r.Context())
		data := lacerateMapDataFromEvents(r.Context(), db)
		ctx := context.WithValue(r.Context(), ctxKeyLacerateMapData, data)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func init() {
	lago.RegistryView.Register("lacerate.MapView",
		lago.GetPageView("lacerate.MapPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.map.data", lacerateMapLayer{}))
}
