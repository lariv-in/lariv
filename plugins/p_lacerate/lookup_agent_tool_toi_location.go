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

func parseAgentLocationDatetime(ctx context.Context, raw string) (time.Time, error) {
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

// targetOfInterestLocationsMapsByTargetID returns location rows as maps for agent payloads (no coordinates).
func targetOfInterestLocationsMapsByTargetID(ctx context.Context, db *gorm.DB, targetIDs []uint) map[uint][]map[string]any {
	out := make(map[uint][]map[string]any)
	if db == nil || db.Name() != "postgres" || len(targetIDs) == 0 {
		return out
	}
	var locs []TargetOfInterestLocation
	if err := db.WithContext(ctx).Preload("Intel").Where("target_of_interest_id IN ?", targetIDs).Order("datetime DESC, id DESC").Find(&locs).Error; err != nil {
		slog.Error("lacerate: load TOI locations for get_relevant_targets_of_interest", "error", err)
		return out
	}
	for _, l := range locs {
		snippet := strings.TrimSpace(l.Intel.Content)
		if len(snippet) > 280 {
			snippet = snippet[:277] + "..."
		}
		out[l.TargetOfInterestID] = append(out[l.TargetOfInterestID], map[string]any{
			"id":            l.ID,
			"intel_id":      l.IntelID,
			"intel_snippet": snippet,
			"datetime":      l.Datetime.UTC().Format(time.RFC3339),
			"address":       l.Address,
		})
	}
	return out
}

type attachTargetOfInterestLocationTool struct{}

func (attachTargetOfInterestLocationTool) Name() string { return "attach_target_of_interest_location" }

func (attachTargetOfInterestLocationTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name: "attach_target_of_interest_location",
		Description: "Add a dated address entry to an existing target of interest, linked to the intel row that supports this geolocation. The address string is sent to the Google Geocoding API to resolve coordinates; " +
			"coordinates are stored but never returned to you—only use address and datetime in tool arguments and in get_relevant_targets_of_interest results.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"target_of_interest_id": map[string]any{"type": "integer", "description": "Target of interest row ID."},
				"intel_id":              map[string]any{"type": "integer", "description": "Intel row ID this location is evidence for (required)."},
				"datetime":              map[string]any{"type": "string", "description": "When this location applies, RFC3339 preferred."},
				"address": map[string]any{
					"type":        "string",
					"description": "Full postal or street address; this exact string is used for Google Geocoding to set the stored map point.",
				},
			},
			"required": []string{"target_of_interest_id", "intel_id", "datetime", "address"},
		},
	}
}

func (attachTargetOfInterestLocationTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	if r.db.Name() != "postgres" {
		err := fmt.Errorf("target of interest locations require PostgreSQL with PostGIS")
		slog.Warn("lacerate: lookup tool attach_target_of_interest_location", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	var p attachTargetOfInterestLocationArgs
	if err := unmarshalToolArgs(args, &p); err != nil {
		slog.Warn("lacerate: lookup tool attach_target_of_interest_location", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if p.TargetOfInterestID == 0 {
		err := fmt.Errorf("target_of_interest_id is required")
		slog.Warn("lacerate: lookup tool attach_target_of_interest_location", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if p.IntelID == 0 {
		err := fmt.Errorf("intel_id is required")
		slog.Warn("lacerate: lookup tool attach_target_of_interest_location", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	addr := strings.TrimSpace(p.Address)
	if addr == "" {
		err := fmt.Errorf("address is required")
		slog.Warn("lacerate: lookup tool attach_target_of_interest_location", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	dt, err := parseAgentLocationDatetime(ctx, p.Datetime)
	if err != nil {
		slog.Warn("lacerate: lookup tool attach_target_of_interest_location", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	var toi TargetOfInterest
	if err := r.db.WithContext(ctx).First(&toi, p.TargetOfInterestID).Error; err != nil {
		slog.Error("lacerate: lookup tool attach_target_of_interest_location load toi", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	var intel Intel
	if err := r.db.WithContext(ctx).First(&intel, p.IntelID).Error; err != nil {
		slog.Error("lacerate: lookup tool attach_target_of_interest_location load intel", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	lat, lng, err := googleGeocodeAddress(ctx, Config.GoogleGeocoding.APIKey, addr)
	if err != nil {
		slog.Warn("lacerate: lookup tool attach_target_of_interest_location geocode", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	ewkbBytes, err := geomPoint4326EWKB(lng, lat)
	if err != nil {
		slog.Warn("lacerate: lookup tool attach_target_of_interest_location ewkb", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	loc, err := insertTargetOfInterestLocation(ctx, r.db, toi.ID, intel.ID, dt, addr, ewkbBytes)
	if err != nil {
		slog.Error("lacerate: lookup tool attach_target_of_interest_location create", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{
		"id":                    loc.ID,
		"target_of_interest_id": loc.TargetOfInterestID,
		"intel_id":              loc.IntelID,
		"datetime":              loc.Datetime.UTC().Format(time.RFC3339),
		"address":               loc.Address,
	}, nil
}

type removeTargetOfInterestLocationTool struct{}

func (removeTargetOfInterestLocationTool) Name() string { return "remove_target_of_interest_location" }

func (removeTargetOfInterestLocationTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "remove_target_of_interest_location",
		Description: "Delete a target-of-interest location row by its id (as returned under locations in get_relevant_targets_of_interest).",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "integer", "description": "TargetOfInterestLocation row ID."},
			},
			"required": []string{"id"},
		},
	}
}

func (removeTargetOfInterestLocationTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	if r.db.Name() != "postgres" {
		err := fmt.Errorf("target of interest locations require PostgreSQL with PostGIS")
		slog.Warn("lacerate: lookup tool remove_target_of_interest_location", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	var p removeTargetOfInterestLocationArgs
	if err := unmarshalToolArgs(args, &p); err != nil {
		slog.Warn("lacerate: lookup tool remove_target_of_interest_location", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if p.ID == 0 {
		err := fmt.Errorf("id is required")
		slog.Warn("lacerate: lookup tool remove_target_of_interest_location", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	res := r.db.WithContext(ctx).Delete(&TargetOfInterestLocation{}, p.ID)
	if res.Error != nil {
		slog.Error("lacerate: lookup tool remove_target_of_interest_location", "error", res.Error, "lookup_id", r.lookupID)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		err := fmt.Errorf("no location row with id %d", p.ID)
		slog.Warn("lacerate: lookup tool remove_target_of_interest_location", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{"id": p.ID, "removed": true}, nil
}
