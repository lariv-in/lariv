package p_seer_assistant

import (
	"github.com/lariv-in/lago/lago"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SeerAssistantSession groups chat rows; [NextTranscriptOrdinal] allocates a single
// ordering across the per-kind message tables.
type SeerAssistantSession struct {
	gorm.Model

	Title                 string `gorm:"not null;default:''"`
	UserID                *uint  `gorm:"index"`
	NextTranscriptOrdinal uint   `gorm:"not null;default:0"`
}

func (SeerAssistantSession) TableName() string {
	return "seer_assistant_sessions"
}

type SeerAssistantUserMessage struct {
	gorm.Model

	SessionID uint   `gorm:"not null;index"`
	Ordinal   uint   `gorm:"not null;index"`
	Body      string `gorm:"type:text;not null"`
}

func (SeerAssistantUserMessage) TableName() string {
	return "seer_assistant_user_messages"
}

type SeerAssistantAssistantMessage struct {
	gorm.Model

	SessionID uint   `gorm:"not null;index"`
	Ordinal   uint   `gorm:"not null;index"`
	Body      string `gorm:"type:text;not null"`
}

func (SeerAssistantAssistantMessage) TableName() string {
	return "seer_assistant_assistant_messages"
}

type SeerAssistantToolCall struct {
	gorm.Model

	SessionID uint           `gorm:"not null;index"`
	Ordinal   uint           `gorm:"not null;index"`
	Name      string         `gorm:"not null;default:''"`
	Arguments datatypes.JSON `gorm:"type:json"`
}

func (SeerAssistantToolCall) TableName() string {
	return "seer_assistant_tool_calls"
}

type SeerAssistantToolResult struct {
	gorm.Model

	ToolCallID uint   `gorm:"not null;index"`
	SessionID  uint   `gorm:"not null;index"`
	Ordinal    uint   `gorm:"not null;index"`
	Result     string `gorm:"type:text;not null"`
	Error      string `gorm:"not null;default:''"`
}

func (SeerAssistantToolResult) TableName() string {
	return "seer_assistant_tool_results"
}

// AssistantChatTurn is one model message (same shape as p_google_genai.ChatTurn).
type AssistantChatTurn struct {
	Role    string
	Content string
}

func init() {
	lago.OnDBInit("p_seer_assistant.models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[SeerAssistantSession](db)
		lago.RegisterModel[SeerAssistantUserMessage](db)
		lago.RegisterModel[SeerAssistantAssistantMessage](db)
		lago.RegisterModel[SeerAssistantToolCall](db)
		lago.RegisterModel[SeerAssistantToolResult](db)
		return db
	})
}
