package p_llm_assistant

import (
	"context"
	"strings"

	"google.golang.org/genai"
	"gorm.io/gorm"
)

func CreateSession(ctx context.Context, db *gorm.DB, userID uint) (LlmAssistantSession, error) {
	session := LlmAssistantSession{UserID: userID}
	if err := db.WithContext(ctx).Create(&session).Error; err != nil {
		return LlmAssistantSession{}, err
	}
	return session, nil
}

func LoadSessionContents(ctx context.Context, db *gorm.DB, sessionID uint) ([]*genai.Content, error) {
	var messages []LlmAssistantSessionMessage
	if err := db.WithContext(ctx).
		Where("llm_assistant_session_id = ?", sessionID).
		Order("id ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	contents := make([]*genai.Content, 0, len(messages))
	for _, message := range messages {
		content, err := message.LoadContent(ctx)
		if err != nil {
			return nil, err
		}
		sanitizeContentPartsForGenaiChat(content)
		contents = append(contents, content)
	}
	return contents, nil
}

func assistantTranscriptTurnKind(c *genai.Content) string {
	if c == nil {
		return "user"
	}
	r := strings.ToLower(strings.TrimSpace(c.Role))
	switch r {
	case string(genai.RoleModel), "assistant":
		return "assistant"
	case "tool":
		return "tool"
	default:
		if assistantContentHasToolResponseParts(c) {
			return "tool"
		}
		return "user"
	}
}

func assistantContentHasToolResponseParts(c *genai.Content) bool {
	if c == nil {
		return false
	}
	for _, p := range c.Parts {
		if p == nil {
			continue
		}
		if p.FunctionResponse != nil || p.ToolResponse != nil {
			return true
		}
	}
	return false
}
