package p_seer_gdelt

import (
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

const EventsTable = "seer_gdelt_events"

// Event stores one fetched GDELT event row using the daily updates schema, including SOURCEURL.
type Event struct {
	gorm.Model

	GlobalEventID uint64 `gorm:"not null;uniqueIndex"`
	SQLDate       int    `gorm:"index"`
	MonthYear     string `gorm:"size:6"`
	Year          string `gorm:"size:4"`
	FractionDate  float64

	Actor1Code           string `gorm:"size:32"`
	Actor1Name           string `gorm:"size:255"`
	Actor1CountryCode    string `gorm:"size:8"`
	Actor1KnownGroupCode string `gorm:"size:32"`
	Actor1EthnicCode     string `gorm:"size:32"`
	Actor1Religion1Code  string `gorm:"size:32"`
	Actor1Religion2Code  string `gorm:"size:32"`
	Actor1Type1Code      string `gorm:"size:32"`
	Actor1Type2Code      string `gorm:"size:32"`
	Actor1Type3Code      string `gorm:"size:32"`

	Actor2Code           string `gorm:"size:32"`
	Actor2Name           string `gorm:"size:255"`
	Actor2CountryCode    string `gorm:"size:8"`
	Actor2KnownGroupCode string `gorm:"size:32"`
	Actor2EthnicCode     string `gorm:"size:32"`
	Actor2Religion1Code  string `gorm:"size:32"`
	Actor2Religion2Code  string `gorm:"size:32"`
	Actor2Type1Code      string `gorm:"size:32"`
	Actor2Type2Code      string `gorm:"size:32"`
	Actor2Type3Code      string `gorm:"size:32"`

	IsRootEvent    int
	EventCode      string `gorm:"size:8"`
	EventBaseCode  string `gorm:"size:8"`
	EventRootCode  string `gorm:"size:8"`
	QuadClass      int
	GoldsteinScale float64
	NumMentions    int
	NumSources     int
	NumArticles    int
	AvgTone        float64

	Actor1GeoType        int
	Actor1GeoFullName    string `gorm:"size:255"`
	Actor1GeoCountryCode string `gorm:"size:8"`
	Actor1GeoADM1Code    string `gorm:"size:32"`
	Actor1GeoADM2Code    string `gorm:"size:32"`
	Actor1GeoLat         float64
	Actor1GeoLong        float64
	Actor1GeoFeatureID   string `gorm:"size:64"`

	Actor2GeoType        int
	Actor2GeoFullName    string `gorm:"size:255"`
	Actor2GeoCountryCode string `gorm:"size:8"`
	Actor2GeoADM1Code    string `gorm:"size:32"`
	Actor2GeoADM2Code    string `gorm:"size:32"`
	Actor2GeoLat         float64
	Actor2GeoLong        float64
	Actor2GeoFeatureID   string `gorm:"size:64"`

	ActionGeoType        int
	ActionGeoFullName    string `gorm:"size:255"`
	ActionGeoCountryCode string `gorm:"size:8"`
	ActionGeoADM1Code    string `gorm:"size:32"`
	ActionGeoADM2Code    string `gorm:"size:32"`
	ActionGeoPoint       lago.PGPoint `gorm:"type:point"`
	ActionGeoLat         float64      `gorm:"-"` // form roundtrip; persisted via [ActionGeoPoint]
	ActionGeoLong        float64      `gorm:"-"`
	ActionGeoFeatureID   string `gorm:"size:64"`

	DateAdded int64
	SourceURL string `gorm:"size:1024"`
}

func (Event) TableName() string {
	return EventsTable
}

func (e *Event) AfterCreate(_ *gorm.DB) error {
	EnqueueEventSourceURLForWebsiteScrape(e.SourceURL)
	return nil
}

func (e *Event) AfterFind(_ *gorm.DB) error {
	e.syncActionGeoFloatsFromPoint()
	return nil
}

func (e *Event) BeforeSave(_ *gorm.DB) error {
	if gdeltValidLatLng(e.ActionGeoLat, e.ActionGeoLong) {
		e.ActionGeoPoint = lago.NewPGPoint(e.ActionGeoLong, e.ActionGeoLat)
	} else {
		e.ActionGeoPoint = lago.PGPoint{}
	}
	return nil
}

func (e *Event) syncActionGeoFloatsFromPoint() {
	if e.ActionGeoPoint.Valid {
		e.ActionGeoLat = e.ActionGeoPoint.P.Y
		e.ActionGeoLong = e.ActionGeoPoint.P.X
	} else {
		e.ActionGeoLat = 0
		e.ActionGeoLong = 0
	}
}

func init() {
	lago.OnDBInit("p_seer_gdelt.models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[Event](db)
		return db
	})
}
