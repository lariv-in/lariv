package p_seer_deepsearch

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

const (
	DeepSearchesTable   = "seer_deep_searches"
	DeepSearchLogsTable = "seer_deep_search_logs"
)

// Status values persisted on [DeepSearch.Status] (Key in [DeepSearchStatusChoices]).
const (
	DeepSearchStatusPending          = "pending"
	DeepSearchStatusRunning          = "running"
	DeepSearchStatusExpandingQueries = "expanding_queries"
	DeepSearchStatusSearching        = "searching"
	DeepSearchStatusScraping         = "scraping"
	DeepSearchStatusIngestingIntel   = "ingesting_intel"
	DeepSearchStatusReporting        = "reporting"
	DeepSearchStatusDone             = "done"
	DeepSearchStatusFailed           = "failed"
	DeepSearchStatusCancelled        = "cancelled"
)

// DeepSearchStatusChoices defines UI labels for [DeepSearch.Status] (Caveats: choice fields).
var DeepSearchStatusChoices = []registry.Pair[string, string]{
	{Key: DeepSearchStatusPending, Value: "Pending"},
	{Key: DeepSearchStatusRunning, Value: "Running"},
	{Key: DeepSearchStatusExpandingQueries, Value: "Expanding queries"},
	{Key: DeepSearchStatusSearching, Value: "Searching the web"},
	{Key: DeepSearchStatusScraping, Value: "Scraping pages"},
	{Key: DeepSearchStatusIngestingIntel, Value: "Adding to Intel"},
	{Key: DeepSearchStatusReporting, Value: "Writing report"},
	{Key: DeepSearchStatusDone, Value: "Done"},
	{Key: DeepSearchStatusFailed, Value: "Failed"},
	{Key: DeepSearchStatusCancelled, Value: "Stopped"},
}

// Log line kinds for [DeepSearchLog.Kind] (Keys in [DeepSearchLogKindChoices]).
const (
	DeepSearchLogKindInfo              = "info"
	DeepSearchLogKindError             = "error"
	DeepSearchLogKindQueriesGenerated  = "queries_generated"
	DeepSearchLogKindSearchPerformed   = "search_performed"
	DeepSearchLogKindWebsiteFetched    = "website_fetched"
	DeepSearchLogKindIntelCreated      = "intel_created"
	DeepSearchLogKindIntelUnchanged    = "intel_unchanged"
	DeepSearchLogKindIntelCreateFailed = "intel_create_failed"
	DeepSearchLogKindReportLlm         = "report_llm"
)

// DeepSearchLogKindChoices labels [DeepSearchLog.Kind] in the UI (Caveats: choice fields).
var DeepSearchLogKindChoices = []registry.Pair[string, string]{
	{Key: DeepSearchLogKindInfo, Value: "Info"},
	{Key: DeepSearchLogKindError, Value: "Error"},
	{Key: DeepSearchLogKindQueriesGenerated, Value: "Queries generated"},
	{Key: DeepSearchLogKindSearchPerformed, Value: "Search performed"},
	{Key: DeepSearchLogKindWebsiteFetched, Value: "Website"},
	{Key: DeepSearchLogKindIntelCreated, Value: "Intel created"},
	{Key: DeepSearchLogKindIntelUnchanged, Value: "Intel unchanged"},
	{Key: DeepSearchLogKindIntelCreateFailed, Value: "Intel failed"},
	{Key: DeepSearchLogKindReportLlm, Value: "Report (LLM)"},
}

// DeepSearchLog is one append-only audit line for a [DeepSearch] run.
type DeepSearchLog struct {
	gorm.Model

	DeepSearchID uint        `gorm:"not null;index"`
	DeepSearch   *DeepSearch `gorm:"foreignKey:DeepSearchID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Kind         string      `gorm:"size:48;not null;default:'';index"`
	Message      string      `gorm:"type:text;not null;default:''"`
}

func (DeepSearchLog) TableName() string {
	return DeepSearchLogsTable
}

// DeepSearch is one background research job from a user query through report generation.
type DeepSearch struct {
	gorm.Model

	Query    string `gorm:"type:text;not null;default:''"`
	Status   string `gorm:"size:32;not null;default:'';index"`
	Report   string `gorm:"type:text;not null;default:''"`
	RunError string `gorm:"type:text;not null;default:''"`

	Logs []DeepSearchLog `gorm:"foreignKey:DeepSearchID"`
}

func (DeepSearch) TableName() string {
	return DeepSearchesTable
}

func init() {
	lago.OnDBInit("p_seer_deepsearch.models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[DeepSearch](db)
		lago.RegisterModel[DeepSearchLog](db)
		return db
	})
}
