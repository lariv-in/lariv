package sqlagent

import (
	"log"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

// Message kinds for ConversationMessage.Kind (which payload row exists).
const (
	MessageKindUser  = "user"
	MessageKindAI    = "ai"
	MessageKindTool  = "tool"
	MessageKindError = "error"
)

// AI message streaming status (AIMessage.Status).
const (
	AIStatusPending   = "pending"
	AIStatusStreaming = "streaming"
	AIStatusComplete  = "complete"
	AIStatusError     = "error"
)

// Conversation is one chat thread owned by a user.
type Conversation struct {
	gorm.Model
	Title       string `gorm:"size:250;not null"`
	CreatedByID uint   `gorm:"not null;index"`
	CreatedBy   p_users.User

	Messages []ConversationMessage `gorm:"constraint:OnDelete:CASCADE;"`
}

// ConversationMessage is the ordered timeline envelope; exactly one payload row should exist per Kind.
type ConversationMessage struct {
	gorm.Model
	ConversationID uint         `gorm:"not null;index"`
	Conversation   Conversation `gorm:"constraint:OnDelete:CASCADE;"`
	SortOrder      int          `gorm:"not null;index"`
	Kind           string       `gorm:"size:32;not null;index"`

	UserMessage  *UserMessage  `gorm:"foreignKey:ConversationMessageID"`
	AIMessage    *AIMessage    `gorm:"foreignKey:ConversationMessageID"`
	ToolMessage  *ToolMessage  `gorm:"foreignKey:ConversationMessageID"`
	ErrorMessage *ErrorMessage `gorm:"foreignKey:ConversationMessageID"`
}

// UserMessage is human-authored content.
type UserMessage struct {
	gorm.Model
	ConversationMessageID uint                `gorm:"uniqueIndex;not null"`
	ConversationMessage   ConversationMessage `gorm:"constraint:OnDelete:CASCADE;"`
	Content               string              `gorm:"type:text;not null"`
}

// AIMessage is assistant output (streamed tokens accumulate in Content).
type AIMessage struct {
	gorm.Model
	ConversationMessageID uint                `gorm:"uniqueIndex;not null"`
	ConversationMessage   ConversationMessage `gorm:"constraint:OnDelete:CASCADE;"`
	Content               string              `gorm:"type:text;not null"`
	Status                string              `gorm:"size:32;not null;default:complete"`
}

// ToolMessage is a placeholder for future tool call / result rows.
type ToolMessage struct {
	gorm.Model
	ConversationMessageID uint                `gorm:"uniqueIndex;not null"`
	ConversationMessage   ConversationMessage `gorm:"constraint:OnDelete:CASCADE;"`
	Name                  string              `gorm:"size:250"`
	Detail                string              `gorm:"type:text"`
}

// ErrorMessage is an inline error bubble in the transcript.
type ErrorMessage struct {
	gorm.Model
	ConversationMessageID uint                `gorm:"uniqueIndex;not null"`
	ConversationMessage   ConversationMessage `gorm:"constraint:OnDelete:CASCADE;"`
	Content               string              `gorm:"type:text;not null"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(
			&Conversation{},
			&ConversationMessage{},
			&UserMessage{},
			&AIMessage{},
			&ToolMessage{},
			&ErrorMessage{},
		); err != nil {
			log.Panicf("sqlagent: automigrate: %v", err)
		}
		return d
	})
}
