package p_seer_gdelt

import (
	"maps"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildGDELTBigQuery(t *testing.T) {
	t.Parallel()

	original := *Config
	defer func() { *Config = original }()
	*Config = GDELTConfig{
		DataProjectID:     "gdelt-bq",
		Dataset:           "gdeltv2",
		Table:             "events",
		DefaultMaxRecords: 50,
	}

	sql, params, err := buildGDELTBigQuery(GDELTSearchRequest{
		Query:         "climate change",
		Domain:        "Example.COM",
		ActionCountry: "us",
		MinMentions:   5,
		MaxRecords:    25,
		Sort:          gdeltSortMentionsDesc,
	})
	if err != nil {
		t.Fatalf("build query: %v", err)
	}

	if !strings.Contains(sql, "`gdelt-bq.gdeltv2.events`") {
		t.Fatalf("expected table name in SQL: %q", sql)
	}
	if !strings.Contains(sql, "ORDER BY NumMentions DESC, SQLDATE DESC") {
		t.Fatalf("expected mentions order in SQL: %q", sql)
	}
	if !strings.Contains(sql, "LIMIT 25") {
		t.Fatalf("expected limit in SQL: %q", sql)
	}

	gotParams := map[string]any{}
	for _, param := range params {
		gotParams[param.Name] = param.Value
	}
	wantParams := map[string]any{
		"query":          "climate change",
		"domain":         "example.com",
		"action_country": "US",
		"min_mentions":   int64(5),
		"start_date":     int64(0),
		"end_date":       int64(0),
	}
	if !maps.Equal(gotParams, wantParams) {
		t.Fatalf("params = %#v, want %#v", gotParams, wantParams)
	}
}

func TestParseGDELTSearchRequest(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "/seer-gdelt/search/?Query=energy&Domain=cnn.com&ActionCountry=US&StartDate=2026-04-01&EndDate=2026-04-20&MinMentions=3&MaxRecords=20&Sort=date_desc", nil)
	search, searched, errs := parseGDELTSearchRequest(r)

	if !searched {
		t.Fatalf("expected searched=true")
	}
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if search.Query != "energy" {
		t.Fatalf("query = %q", search.Query)
	}
	if search.Domain != "cnn.com" {
		t.Fatalf("domain = %q", search.Domain)
	}
	if search.ActionCountry != "US" {
		t.Fatalf("action country = %q", search.ActionCountry)
	}
	if search.MinMentions != 3 {
		t.Fatalf("min mentions = %d", search.MinMentions)
	}
	if search.MaxRecords != 20 {
		t.Fatalf("max records = %d", search.MaxRecords)
	}
	if search.Sort != gdeltSortDateDesc {
		t.Fatalf("sort = %q", search.Sort)
	}
	if search.StartDate == nil || search.StartDate.Format("2006-01-02") != "2026-04-01" {
		t.Fatalf("start date = %v", search.StartDate)
	}
	if search.EndDate == nil || search.EndDate.UTC().Format("2006-01-02 15:04:05") != "2026-04-20 23:59:59" {
		t.Fatalf("end date = %v", search.EndDate)
	}
}

func TestParseGDELTSearchRequestRejectsInvalidMaxRecords(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "/seer-gdelt/search/?Query=energy&MaxRecords=999", nil)
	_, searched, errs := parseGDELTSearchRequest(r)

	if !searched {
		t.Fatalf("expected searched=true")
	}
	if errs["MaxRecords"] == nil {
		t.Fatalf("expected max records error, got %v", errs)
	}
}
