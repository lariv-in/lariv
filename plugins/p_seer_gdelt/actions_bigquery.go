package p_seer_gdelt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

const (
	defaultGDELTMaxRecords = 50
	maxGDELTMaxRecords     = 250
	gdeltSortDateDesc      = "date_desc"
	gdeltSortDateAsc       = "date_asc"
	gdeltSortMentionsDesc  = "mentions_desc"
)

type GDELTSearchRequest struct {
	Query         string
	Domain        string
	ActionCountry string
	StartDate     *time.Time
	EndDate       *time.Time
	MinMentions   uint
	MaxRecords    uint
	Sort          string
}

type gdeltBigQueryRow struct {
	// INTEGER in BigQuery; client decodes as int64 (not uint64).
	GlobalEventID int64 `bigquery:"GLOBALEVENTID"`
	SQLDate       int64   `bigquery:"SQLDATE"`
	MonthYear     int64   `bigquery:"MonthYear"`
	Year          int64   `bigquery:"Year"`
	FractionDate  float64 `bigquery:"FractionDate"`

	Actor1Code           string `bigquery:"Actor1Code"`
	Actor1Name           string `bigquery:"Actor1Name"`
	Actor1CountryCode    string `bigquery:"Actor1CountryCode"`
	Actor1KnownGroupCode string `bigquery:"Actor1KnownGroupCode"`
	Actor1EthnicCode     string `bigquery:"Actor1EthnicCode"`
	Actor1Religion1Code  string `bigquery:"Actor1Religion1Code"`
	Actor1Religion2Code  string `bigquery:"Actor1Religion2Code"`
	Actor1Type1Code      string `bigquery:"Actor1Type1Code"`
	Actor1Type2Code      string `bigquery:"Actor1Type2Code"`
	Actor1Type3Code      string `bigquery:"Actor1Type3Code"`

	Actor2Code           string `bigquery:"Actor2Code"`
	Actor2Name           string `bigquery:"Actor2Name"`
	Actor2CountryCode    string `bigquery:"Actor2CountryCode"`
	Actor2KnownGroupCode string `bigquery:"Actor2KnownGroupCode"`
	Actor2EthnicCode     string `bigquery:"Actor2EthnicCode"`
	Actor2Religion1Code  string `bigquery:"Actor2Religion1Code"`
	Actor2Religion2Code  string `bigquery:"Actor2Religion2Code"`
	Actor2Type1Code      string `bigquery:"Actor2Type1Code"`
	Actor2Type2Code      string `bigquery:"Actor2Type2Code"`
	Actor2Type3Code      string `bigquery:"Actor2Type3Code"`

	IsRootEvent    int64   `bigquery:"IsRootEvent"`
	EventCode      string  `bigquery:"EventCode"`
	EventBaseCode  string  `bigquery:"EventBaseCode"`
	EventRootCode  string  `bigquery:"EventRootCode"`
	QuadClass      int64   `bigquery:"QuadClass"`
	GoldsteinScale float64 `bigquery:"GoldsteinScale"`
	NumMentions    int64   `bigquery:"NumMentions"`
	NumSources     int64   `bigquery:"NumSources"`
	NumArticles    int64   `bigquery:"NumArticles"`
	AvgTone        float64 `bigquery:"AvgTone"`

	Actor1GeoType        int64   `bigquery:"Actor1Geo_Type"`
	Actor1GeoFullName    string  `bigquery:"Actor1Geo_FullName"`
	Actor1GeoCountryCode string  `bigquery:"Actor1Geo_CountryCode"`
	Actor1GeoADM1Code    string  `bigquery:"Actor1Geo_ADM1Code"`
	Actor1GeoADM2Code    string  `bigquery:"Actor1Geo_ADM2Code"`
	Actor1GeoLat         float64 `bigquery:"Actor1Geo_Lat"`
	Actor1GeoLong        float64 `bigquery:"Actor1Geo_Long"`
	Actor1GeoFeatureID   string  `bigquery:"Actor1Geo_FeatureID"`

	Actor2GeoType        int64   `bigquery:"Actor2Geo_Type"`
	Actor2GeoFullName    string  `bigquery:"Actor2Geo_FullName"`
	Actor2GeoCountryCode string  `bigquery:"Actor2Geo_CountryCode"`
	Actor2GeoADM1Code    string  `bigquery:"Actor2Geo_ADM1Code"`
	Actor2GeoADM2Code    string  `bigquery:"Actor2Geo_ADM2Code"`
	Actor2GeoLat         float64 `bigquery:"Actor2Geo_Lat"`
	Actor2GeoLong        float64 `bigquery:"Actor2Geo_Long"`
	Actor2GeoFeatureID   string  `bigquery:"Actor2Geo_FeatureID"`

	ActionGeoType        int64   `bigquery:"ActionGeo_Type"`
	ActionGeoFullName    string  `bigquery:"ActionGeo_FullName"`
	ActionGeoCountryCode string  `bigquery:"ActionGeo_CountryCode"`
	ActionGeoADM1Code    string  `bigquery:"ActionGeo_ADM1Code"`
	ActionGeoADM2Code    string  `bigquery:"ActionGeo_ADM2Code"`
	ActionGeoLat         float64 `bigquery:"ActionGeo_Lat"`
	ActionGeoLong        float64 `bigquery:"ActionGeo_Long"`
	ActionGeoFeatureID   string  `bigquery:"ActionGeo_FeatureID"`

	DateAdded int64  `bigquery:"DATEADDED"`
	SourceURL string `bigquery:"SOURCEURL"`
}

func FetchAndStoreGDELTEvents(ctx context.Context, db *gorm.DB, search GDELTSearchRequest) ([]Event, error) {
	if strings.TrimSpace(Config.ProjectID) == "" {
		return nil, fmt.Errorf("configure [Plugins.p_seer_gdelt].projectID to run BigQuery searches")
	}
	var opts []option.ClientOption
	if strings.TrimSpace(Config.CredentialsFile) != "" {
		opts = append(opts, option.WithAuthCredentialsFile(option.ServiceAccount, Config.CredentialsFile))
	}
	client, err := bigquery.NewClient(ctx, Config.ProjectID, opts...)
	if err != nil {
		slog.Error("p_seer_gdelt: failed creating BigQuery client", "project_id", Config.ProjectID, "error", err)
		return nil, err
	}
	defer client.Close()

	sql, params, err := buildGDELTBigQuery(search)
	if err != nil {
		return nil, err
	}
	q := client.Query(sql)
	q.Location = Config.Location
	q.Parameters = params
	q.UseStandardSQL = true

	it, err := q.Read(ctx)
	if err != nil {
		slog.Error("p_seer_gdelt: BigQuery read failed", "error", err)
		return nil, err
	}

	var rows []gdeltBigQueryRow
	for {
		var row gdeltBigQueryRow
		err := it.Next(&row)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			slog.Error("p_seer_gdelt: BigQuery row decode failed", "error", err)
			return nil, err
		}
		rows = append(rows, row)
	}
	return upsertEventsFromBigQuery(ctx, db, rows)
}

func buildGDELTBigQuery(search GDELTSearchRequest) (string, []bigquery.QueryParameter, error) {
	table := gdeltTableName()
	if table == "" {
		return "", nil, fmt.Errorf("GDELT BigQuery table is not configured")
	}
	sql := fmt.Sprintf(`
SELECT
  COALESCE(GLOBALEVENTID, 0) AS GLOBALEVENTID,
  COALESCE(SQLDATE, 0) AS SQLDATE,
  COALESCE(MonthYear, 0) AS MonthYear,
  COALESCE(Year, 0) AS Year,
  COALESCE(FractionDate, 0) AS FractionDate,
  COALESCE(Actor1Code, '') AS Actor1Code,
  COALESCE(Actor1Name, '') AS Actor1Name,
  COALESCE(Actor1CountryCode, '') AS Actor1CountryCode,
  COALESCE(Actor1KnownGroupCode, '') AS Actor1KnownGroupCode,
  COALESCE(Actor1EthnicCode, '') AS Actor1EthnicCode,
  COALESCE(Actor1Religion1Code, '') AS Actor1Religion1Code,
  COALESCE(Actor1Religion2Code, '') AS Actor1Religion2Code,
  COALESCE(Actor1Type1Code, '') AS Actor1Type1Code,
  COALESCE(Actor1Type2Code, '') AS Actor1Type2Code,
  COALESCE(Actor1Type3Code, '') AS Actor1Type3Code,
  COALESCE(Actor2Code, '') AS Actor2Code,
  COALESCE(Actor2Name, '') AS Actor2Name,
  COALESCE(Actor2CountryCode, '') AS Actor2CountryCode,
  COALESCE(Actor2KnownGroupCode, '') AS Actor2KnownGroupCode,
  COALESCE(Actor2EthnicCode, '') AS Actor2EthnicCode,
  COALESCE(Actor2Religion1Code, '') AS Actor2Religion1Code,
  COALESCE(Actor2Religion2Code, '') AS Actor2Religion2Code,
  COALESCE(Actor2Type1Code, '') AS Actor2Type1Code,
  COALESCE(Actor2Type2Code, '') AS Actor2Type2Code,
  COALESCE(Actor2Type3Code, '') AS Actor2Type3Code,
  COALESCE(IsRootEvent, 0) AS IsRootEvent,
  COALESCE(EventCode, '') AS EventCode,
  COALESCE(EventBaseCode, '') AS EventBaseCode,
  COALESCE(EventRootCode, '') AS EventRootCode,
  COALESCE(QuadClass, 0) AS QuadClass,
  COALESCE(GoldsteinScale, 0) AS GoldsteinScale,
  COALESCE(NumMentions, 0) AS NumMentions,
  COALESCE(NumSources, 0) AS NumSources,
  COALESCE(NumArticles, 0) AS NumArticles,
  COALESCE(AvgTone, 0) AS AvgTone,
  COALESCE(Actor1Geo_Type, 0) AS Actor1Geo_Type,
  COALESCE(Actor1Geo_FullName, '') AS Actor1Geo_FullName,
  COALESCE(Actor1Geo_CountryCode, '') AS Actor1Geo_CountryCode,
  COALESCE(Actor1Geo_ADM1Code, '') AS Actor1Geo_ADM1Code,
  COALESCE(Actor1Geo_ADM2Code, '') AS Actor1Geo_ADM2Code,
  COALESCE(Actor1Geo_Lat, 0) AS Actor1Geo_Lat,
  COALESCE(Actor1Geo_Long, 0) AS Actor1Geo_Long,
  COALESCE(CAST(Actor1Geo_FeatureID AS STRING), '') AS Actor1Geo_FeatureID,
  COALESCE(Actor2Geo_Type, 0) AS Actor2Geo_Type,
  COALESCE(Actor2Geo_FullName, '') AS Actor2Geo_FullName,
  COALESCE(Actor2Geo_CountryCode, '') AS Actor2Geo_CountryCode,
  COALESCE(Actor2Geo_ADM1Code, '') AS Actor2Geo_ADM1Code,
  COALESCE(Actor2Geo_ADM2Code, '') AS Actor2Geo_ADM2Code,
  COALESCE(Actor2Geo_Lat, 0) AS Actor2Geo_Lat,
  COALESCE(Actor2Geo_Long, 0) AS Actor2Geo_Long,
  COALESCE(CAST(Actor2Geo_FeatureID AS STRING), '') AS Actor2Geo_FeatureID,
  COALESCE(ActionGeo_Type, 0) AS ActionGeo_Type,
  COALESCE(ActionGeo_FullName, '') AS ActionGeo_FullName,
  COALESCE(ActionGeo_CountryCode, '') AS ActionGeo_CountryCode,
  COALESCE(ActionGeo_ADM1Code, '') AS ActionGeo_ADM1Code,
  COALESCE(ActionGeo_ADM2Code, '') AS ActionGeo_ADM2Code,
  COALESCE(ActionGeo_Lat, 0) AS ActionGeo_Lat,
  COALESCE(ActionGeo_Long, 0) AS ActionGeo_Long,
  COALESCE(CAST(ActionGeo_FeatureID AS STRING), '') AS ActionGeo_FeatureID,
  COALESCE(DATEADDED, 0) AS DATEADDED,
  COALESCE(SOURCEURL, '') AS SOURCEURL
FROM %s
WHERE
  (@query = '' OR STRPOS(LOWER(CONCAT(COALESCE(Actor1Name, ''), ' ', COALESCE(Actor2Name, ''), ' ', COALESCE(ActionGeo_FullName, ''), ' ', COALESCE(SOURCEURL, ''))), LOWER(@query)) > 0)
  AND (@domain = '' OR ENDS_WITH(LOWER(COALESCE(NET.HOST(SOURCEURL), '')), LOWER(@domain)))
  AND (@action_country = '' OR UPPER(COALESCE(ActionGeo_CountryCode, '')) = UPPER(@action_country))
  AND (@min_mentions = 0 OR NumMentions >= @min_mentions)
  AND (@start_date = 0 OR SQLDATE >= @start_date)
  AND (@end_date = 0 OR SQLDATE <= @end_date)
ORDER BY %s
LIMIT %d
`, table, gdeltOrderBy(search.Sort), limitGDELTMaxRecords(search.MaxRecords))

	params := []bigquery.QueryParameter{
		{Name: "query", Value: strings.TrimSpace(search.Query)},
		{Name: "domain", Value: normalizeGDELTDomain(search.Domain)},
		{Name: "action_country", Value: normalizeGDELTCode(search.ActionCountry)},
		{Name: "min_mentions", Value: int64(search.MinMentions)},
		{Name: "start_date", Value: int64(gdeltDateNumber(search.StartDate))},
		{Name: "end_date", Value: int64(gdeltDateNumber(search.EndDate))},
	}
	return sql, params, nil
}

func gdeltTableName() string {
	projectID := strings.TrimSpace(Config.DataProjectID)
	dataset := strings.TrimSpace(Config.Dataset)
	table := strings.TrimSpace(Config.Table)
	if projectID == "" || dataset == "" || table == "" {
		return ""
	}
	return fmt.Sprintf("`%s.%s.%s`", projectID, dataset, table)
}

func limitGDELTMaxRecords(v uint) uint {
	if v == 0 {
		if Config.DefaultMaxRecords > 0 {
			return Config.DefaultMaxRecords
		}
		return defaultGDELTMaxRecords
	}
	if v > maxGDELTMaxRecords {
		return maxGDELTMaxRecords
	}
	return v
}

func gdeltSortOrDefault(s string) string {
	switch strings.TrimSpace(s) {
	case gdeltSortDateAsc, gdeltSortDateDesc, gdeltSortMentionsDesc:
		return strings.TrimSpace(s)
	default:
		return gdeltSortDateDesc
	}
}

func gdeltOrderBy(s string) string {
	switch gdeltSortOrDefault(s) {
	case gdeltSortDateAsc:
		return "SQLDATE ASC, NumMentions DESC"
	case gdeltSortMentionsDesc:
		return "NumMentions DESC, SQLDATE DESC"
	default:
		return "SQLDATE DESC, NumMentions DESC"
	}
}

func normalizeGDELTDomain(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "www.")
	s = strings.TrimSuffix(s, "/")
	return s
}

func normalizeGDELTCode(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

func gdeltDateNumber(t *time.Time) int {
	if t == nil || t.IsZero() {
		return 0
	}
	n, err := strconv.Atoi(t.Format("20060102"))
	if err != nil {
		slog.Warn("p_seer_gdelt: failed converting date to SQLDATE", "value", t, "error", err)
		return 0
	}
	return n
}

// upsertEventsFromBigQuery inserts or updates rows by GlobalEventID. It does not delete other local
// rows, so the Events table accumulates results across searches (subject to BigQuery LIMIT per run).
func upsertEventsFromBigQuery(ctx context.Context, db *gorm.DB, rows []gdeltBigQueryRow) ([]Event, error) {
	if db == nil {
		return nil, fmt.Errorf("p_seer_gdelt: db is nil")
	}
	out := make([]Event, 0, len(rows))
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, row := range rows {
			ev := eventFromBigQueryRow(row)
			var existing Event
			err := tx.Where("global_event_id = ?", ev.GlobalEventID).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&ev).Error; err != nil {
					slog.Error("p_seer_gdelt: insert event failed", "error", err, "global_event_id", ev.GlobalEventID)
					return err
				}
				out = append(out, ev)
				continue
			}
			if err != nil {
				slog.Error("p_seer_gdelt: lookup event failed", "error", err, "global_event_id", ev.GlobalEventID)
				return err
			}
			ev.ID = existing.ID
			ev.CreatedAt = existing.CreatedAt
			if err := tx.Session(&gorm.Session{FullSaveAssociations: false}).Save(&ev).Error; err != nil {
				slog.Error("p_seer_gdelt: update event failed", "error", err, "global_event_id", ev.GlobalEventID)
				return err
			}
			out = append(out, ev)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func eventFromBigQueryRow(row gdeltBigQueryRow) Event {
	return Event{
		GlobalEventID: uint64(row.GlobalEventID),
		SQLDate:       int(row.SQLDate),
		MonthYear:     strconv.FormatInt(row.MonthYear, 10),
		Year:          strconv.FormatInt(row.Year, 10),
		FractionDate:  row.FractionDate,

		Actor1Code:           row.Actor1Code,
		Actor1Name:           row.Actor1Name,
		Actor1CountryCode:    row.Actor1CountryCode,
		Actor1KnownGroupCode: row.Actor1KnownGroupCode,
		Actor1EthnicCode:     row.Actor1EthnicCode,
		Actor1Religion1Code:  row.Actor1Religion1Code,
		Actor1Religion2Code:  row.Actor1Religion2Code,
		Actor1Type1Code:      row.Actor1Type1Code,
		Actor1Type2Code:      row.Actor1Type2Code,
		Actor1Type3Code:      row.Actor1Type3Code,

		Actor2Code:           row.Actor2Code,
		Actor2Name:           row.Actor2Name,
		Actor2CountryCode:    row.Actor2CountryCode,
		Actor2KnownGroupCode: row.Actor2KnownGroupCode,
		Actor2EthnicCode:     row.Actor2EthnicCode,
		Actor2Religion1Code:  row.Actor2Religion1Code,
		Actor2Religion2Code:  row.Actor2Religion2Code,
		Actor2Type1Code:      row.Actor2Type1Code,
		Actor2Type2Code:      row.Actor2Type2Code,
		Actor2Type3Code:      row.Actor2Type3Code,

		IsRootEvent:    int(row.IsRootEvent),
		EventCode:      row.EventCode,
		EventBaseCode:  row.EventBaseCode,
		EventRootCode:  row.EventRootCode,
		QuadClass:      int(row.QuadClass),
		GoldsteinScale: row.GoldsteinScale,
		NumMentions:    int(row.NumMentions),
		NumSources:     int(row.NumSources),
		NumArticles:    int(row.NumArticles),
		AvgTone:        row.AvgTone,

		Actor1GeoType:        int(row.Actor1GeoType),
		Actor1GeoFullName:    row.Actor1GeoFullName,
		Actor1GeoCountryCode: row.Actor1GeoCountryCode,
		Actor1GeoADM1Code:    row.Actor1GeoADM1Code,
		Actor1GeoADM2Code:    row.Actor1GeoADM2Code,
		Actor1GeoLat:         row.Actor1GeoLat,
		Actor1GeoLong:        row.Actor1GeoLong,
		Actor1GeoFeatureID:   row.Actor1GeoFeatureID,

		Actor2GeoType:        int(row.Actor2GeoType),
		Actor2GeoFullName:    row.Actor2GeoFullName,
		Actor2GeoCountryCode: row.Actor2GeoCountryCode,
		Actor2GeoADM1Code:    row.Actor2GeoADM1Code,
		Actor2GeoADM2Code:    row.Actor2GeoADM2Code,
		Actor2GeoLat:         row.Actor2GeoLat,
		Actor2GeoLong:        row.Actor2GeoLong,
		Actor2GeoFeatureID:   row.Actor2GeoFeatureID,

		ActionGeoType:        int(row.ActionGeoType),
		ActionGeoFullName:    row.ActionGeoFullName,
		ActionGeoCountryCode: row.ActionGeoCountryCode,
		ActionGeoADM1Code:    row.ActionGeoADM1Code,
		ActionGeoADM2Code:    row.ActionGeoADM2Code,
		ActionGeoLat:         row.ActionGeoLat,
		ActionGeoLong:        row.ActionGeoLong,
		ActionGeoFeatureID:   row.ActionGeoFeatureID,

		DateAdded: row.DateAdded,
		SourceURL: row.SourceURL,
	}
}
