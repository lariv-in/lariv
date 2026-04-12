package p_lacerate

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/genai"
	"gorm.io/gorm"
)

func parseAgentEventDatetime(ctx context.Context, raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("datetime is required")
	}
	tz, _ := ctx.Value("$tz").(*time.Location)
	if tz == nil {
		tz = time.UTC
	}
	dt, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		dt, err = time.ParseInLocation("2006-01-02T15:04", raw, tz)
		if err != nil {
			return time.Time{}, fmt.Errorf("datetime must be RFC3339 (or local YYYY-MM-DDTHH:MM)")
		}
	}
	return dt, nil
}

// eventsMapsByIntelID returns event rows as maps for agent payloads (no coordinates), keyed by intel id.
func eventsMapsByIntelID(ctx context.Context, db *gorm.DB, intelIDs []uint) map[uint][]map[string]any {
	out := make(map[uint][]map[string]any)
	if db == nil || db.Name() != "postgres" || len(intelIDs) == 0 {
		return out
	}
	var evs []Event
	if err := db.WithContext(ctx).Where("intel_id IN ?", intelIDs).Order("intel_id ASC, datetime DESC, id DESC").Find(&evs).Error; err != nil {
		slog.Error("lacerate: load events for intel ids", "error", err)
		return out
	}
	for _, e := range evs {
		out[e.IntelID] = append(out[e.IntelID], map[string]any{
			"id":       e.ID,
			"datetime": e.Datetime.UTC().Format(time.RFC3339),
			"address":  strings.TrimSpace(e.Address),
		})
	}
	return out
}

type attachEventTool struct{}

func (attachEventTool) Name() string { return "attach_event" }

func (attachEventTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name: "attach_event",
		Description: "Add a dated geocoded address (event) linked to an intel row. The address string is sent to the Google Geocoding API to resolve coordinates; " +
			"coordinates are stored but never returned to you—only use address and datetime in tool arguments and in get_relevant_intel event lists.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"intel_id": map[string]any{"type": "integer", "description": "Intel row ID this event is tied to (required)."},
				"datetime": map[string]any{"type": "string", "description": "When this location applies, RFC3339 preferred."},
				"address": map[string]any{
					"type":        "string",
					"description": "Full postal or street address; this exact string is used for Google Geocoding to set the stored map point.",
				},
			},
			"required": []string{"intel_id", "datetime", "address"},
		},
	}
}

func (attachEventTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	if r.db.Name() != "postgres" {
		err := fmt.Errorf("events require PostgreSQL with PostGIS")
		slog.Warn("lacerate: lookup tool attach_event", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	var p attachEventArgs
	if err := unmarshalToolArgs(args, &p); err != nil {
		slog.Warn("lacerate: lookup tool attach_event", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if p.IntelID == 0 {
		err := fmt.Errorf("intel_id is required")
		slog.Warn("lacerate: lookup tool attach_event", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	addr := strings.TrimSpace(p.Address)
	if addr == "" {
		err := fmt.Errorf("address is required")
		slog.Warn("lacerate: lookup tool attach_event", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	dt, err := parseAgentEventDatetime(ctx, p.Datetime)
	if err != nil {
		slog.Warn("lacerate: lookup tool attach_event", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	var intel Intel
	if err := r.db.WithContext(ctx).First(&intel, p.IntelID).Error; err != nil {
		slog.Error("lacerate: lookup tool attach_event load intel", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	lat, lng, err := googleGeocodeAddress(ctx, Config.GoogleGeocoding.APIKey, addr)
	if err != nil {
		slog.Warn("lacerate: lookup tool attach_event geocode", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	ewkbBytes, err := geomPoint4326EWKB(lng, lat)
	if err != nil {
		slog.Warn("lacerate: lookup tool attach_event ewkb", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	ev, err := insertEvent(ctx, r.db, intel.ID, dt, addr, ewkbBytes)
	if err != nil {
		slog.Error("lacerate: lookup tool attach_event create", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{
		"id":       ev.ID,
		"intel_id": ev.IntelID,
		"datetime": ev.Datetime.UTC().Format(time.RFC3339),
		"address":  ev.Address,
	}, nil
}

type removeEventTool struct{}

func (removeEventTool) Name() string { return "remove_event" }

func (removeEventTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "remove_event",
		Description: "Delete an event row by its id (as returned under events in get_relevant_intel for that intel).",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "integer", "description": "Event row ID."},
			},
			"required": []string{"id"},
		},
	}
}

func (removeEventTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	if r.db.Name() != "postgres" {
		err := fmt.Errorf("events require PostgreSQL with PostGIS")
		slog.Warn("lacerate: lookup tool remove_event", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	var p removeEventArgs
	if err := unmarshalToolArgs(args, &p); err != nil {
		slog.Warn("lacerate: lookup tool remove_event", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if p.ID == 0 {
		err := fmt.Errorf("id is required")
		slog.Warn("lacerate: lookup tool remove_event", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	res := r.db.WithContext(ctx).Delete(&Event{}, p.ID)
	if res.Error != nil {
		slog.Error("lacerate: lookup tool remove_event", "error", res.Error, "lookup_id", r.lookupID)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		err := fmt.Errorf("no event row with id %d", p.ID)
		slog.Warn("lacerate: lookup tool remove_event", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{"id": p.ID, "removed": true}, nil
}
