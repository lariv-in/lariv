package p_lacerate

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

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

// LookupThought stores free-form reasoning tied to a log entry (Kind "thought").
type LookupThought struct {
	gorm.Model
	LookupLogEntryID uint           `gorm:"not null;uniqueIndex"`
	LookupLogEntry   LookupLogEntry `gorm:"foreignKey:LookupLogEntryID;constraint:OnDelete:CASCADE"`
	Text             string         `gorm:"type:text;not null"`
}

// LookupText stores plain text output (Kind "text").
type LookupText struct {
	gorm.Model
	LookupLogEntryID uint           `gorm:"not null;uniqueIndex"`
	LookupLogEntry   LookupLogEntry `gorm:"foreignKey:LookupLogEntryID;constraint:OnDelete:CASCADE"`
	Text             string         `gorm:"type:text;not null"`
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

// LookupToolError records a tool execution failure (Kind "tool_error").
type LookupToolError struct {
	gorm.Model
	LookupLogEntryID uint           `gorm:"not null;uniqueIndex"`
	LookupLogEntry   LookupLogEntry `gorm:"foreignKey:LookupLogEntryID;constraint:OnDelete:CASCADE"`
	ToolName         string         `gorm:"not null;size:128"`
	Message          string         `gorm:"type:text;not null"`
	Detail           datatypes.JSON
}

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
