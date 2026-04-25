package p_seer_assistant

import (
	"context"
	"strings"

	"google.golang.org/genai"
	"gorm.io/gorm"
)

func CreateSession(ctx context.Context, db *gorm.DB, userID uint) (SeerAssistantSession, error) {
	session := SeerAssistantSession{UserID: userID}
	if err := db.WithContext(ctx).Create(&session).Error; err != nil {
		return SeerAssistantSession{}, err
	}
	return session, nil
}

func LoadSessionContents(ctx context.Context, db *gorm.DB, sessionID uint) ([]*genai.Content, error) {
	var messages []SeerAssistantSessionMessage
	if err := db.WithContext(ctx).
		Where("seer_assistant_session_id = ?", sessionID).
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
		contents = append(contents, content)
	}
	return contents, nil
}

// assistantTranscriptTurnKind picks bubble column for transcript UI ("user" | "assistant" | "tool").
// Tool results are stored as role user with function/tool response parts (see actions_chat_llm).
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
