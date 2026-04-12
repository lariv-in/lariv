package p_lacerate

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// LookupReportTouchActionChoices labels create/edit actions derived from lookup tool calls (see [buildLookupTouchedReportDisplays]).
var LookupReportTouchActionChoices = []registry.Pair[string, string]{
	{Key: "create", Value: "Created"},
	{Key: "edit", Value: "Edited"},
}

// LookupLogEntryKindChoices maps persisted Kind values to UI labels for [LookupLogEntry].
var LookupLogEntryKindChoices = []registry.Pair[string, string]{
	{Key: "thought", Value: "Thought"},
	{Key: "text", Value: "Text"},
	{Key: "tool_call", Value: "Tool Call"},
	{Key: "tool_error", Value: "Tool Error"},
}

// LookupLogEntry is one timeline row for an automated lookup run ([runLookupAgent]).
type LookupLogEntry struct {
	gorm.Model
	LookupID uint   `gorm:"not null;index"`
	Lookup   Lookup `gorm:"foreignKey:LookupID;constraint:OnDelete:CASCADE"`
	Kind     string `gorm:"not null;size:32"`
}

// LookupLogEntryData is persisted payload for a [LookupLogEntry] (thought, text, tool_call, tool_error).
type LookupLogEntryData interface {
	Kind() string
	// SetLookup binds the parent [LookupLogEntry] (id and association) before insert.
	SetLookup(entry *LookupLogEntry)
}

func withLookupLogEntryTx(db *gorm.DB, lookupID uint, kind string, fn func(tx *gorm.DB, entry *LookupLogEntry) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		entry := LookupLogEntry{LookupID: lookupID, Kind: kind}
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}
		return fn(tx, &entry)
	})
}

// CreateLookupLogEntryData inserts a [LookupLogEntry] for lookupID and row in one transaction.
// row must be a pointer type that implements [LookupLogEntryData]; [SetLookup] runs before insert.
func CreateLookupLogEntryData[T LookupLogEntryData](db *gorm.DB, lookupID uint, row T) error {
	return withLookupLogEntryTx(db, lookupID, row.Kind(), func(tx *gorm.DB, entry *LookupLogEntry) error {
		row.SetLookup(entry)
		return tx.Create(row).Error
	})
}

// LookupThought stores free-form reasoning tied to a log entry (Kind "thought").
type LookupThought struct {
	gorm.Model
	LookupLogEntryID uint           `gorm:"not null;uniqueIndex"`
	LookupLogEntry   LookupLogEntry `gorm:"foreignKey:LookupLogEntryID;constraint:OnDelete:CASCADE"`
	Text             string         `gorm:"type:text;not null"`
}

func (LookupThought) Kind() string { return "thought" }

func (t *LookupThought) SetLookup(entry *LookupLogEntry) {
	if entry == nil {
		return
	}
	t.LookupLogEntryID = entry.ID
	t.LookupLogEntry = *entry
}

// LookupText stores plain text output (Kind "text").
type LookupText struct {
	gorm.Model
	LookupLogEntryID uint           `gorm:"not null;uniqueIndex"`
	LookupLogEntry   LookupLogEntry `gorm:"foreignKey:LookupLogEntryID;constraint:OnDelete:CASCADE"`
	Text             string         `gorm:"type:text;not null"`
}

func (LookupText) Kind() string { return "text" }

func (t *LookupText) SetLookup(entry *LookupLogEntry) {
	if entry == nil {
		return
	}
	t.LookupLogEntryID = entry.ID
	t.LookupLogEntry = *entry
}

// LookupToolCall records a model tool invocation (Kind "tool_call").
type LookupToolCall struct {
	gorm.Model
	LookupLogEntryID uint           `gorm:"not null;uniqueIndex"`
	LookupLogEntry   LookupLogEntry `gorm:"foreignKey:LookupLogEntryID;constraint:OnDelete:CASCADE"`
	Name             string         `gorm:"not null;size:128"`
	Arguments        datatypes.JSON
	Result           datatypes.JSON
}

func (LookupToolCall) Kind() string { return "tool_call" }

func (t *LookupToolCall) SetLookup(entry *LookupLogEntry) {
	if entry == nil {
		return
	}
	t.LookupLogEntryID = entry.ID
	t.LookupLogEntry = *entry
}

// LookupToolError records a tool execution failure (Kind "tool_error").
type LookupToolError struct {
	gorm.Model
	LookupLogEntryID uint           `gorm:"not null;uniqueIndex"`
	LookupLogEntry   LookupLogEntry `gorm:"foreignKey:LookupLogEntryID;constraint:OnDelete:CASCADE"`
	ToolName         string         `gorm:"not null;size:128"`
	Message          string         `gorm:"type:text;not null"`
	Detail           datatypes.JSON
}

func (LookupToolError) Kind() string { return "tool_error" }

func (t *LookupToolError) SetLookup(entry *LookupLogEntry) {
	if entry == nil {
		return
	}
	t.LookupLogEntryID = entry.ID
	t.LookupLogEntry = *entry
}

var (
	_ LookupLogEntryData = (*LookupThought)(nil)
	_ LookupLogEntryData = (*LookupText)(nil)
	_ LookupLogEntryData = (*LookupToolCall)(nil)
	_ LookupLogEntryData = (*LookupToolError)(nil)
)

func init() {
	lago.OnDBInit("p_lacerate.lookup_log_models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[LookupLogEntry](db)
		lago.RegisterModel[LookupThought](db)
		lago.RegisterModel[LookupText](db)
		lago.RegisterModel[LookupToolCall](db)
		lago.RegisterModel[LookupToolError](db)
		return db
	})
}
