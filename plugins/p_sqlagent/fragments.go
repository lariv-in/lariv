package sqlagent

import (
	"context"
	"fmt"
	"html"
	"strings"

	"gorm.io/gorm"
)

// RenderMessageBubble returns a single chat bubble as HTML (no full document).
func RenderMessageBubble(msg ConversationMessage) string {
	if isRegistrySchemaToolMessage(&msg) {
		return ""
	}
	if msg.Kind == MessageKindUser && msg.UserMessage != nil && isRegistrySchemaBootstrapUserContent(msg.UserMessage.Content) {
		return ""
	}
	id := fmt.Sprintf("sqlagent-msg-%d", msg.ID)
	switch msg.Kind {
	case MessageKindUser:
		if msg.UserMessage == nil {
			return ""
		}
		return renderUserBubble(id, msg.UserMessage.Content, false)
	case MessageKindAI:
		if msg.AIMessage == nil {
			return ""
		}
		return renderAIBubble(id, msg.AIMessage.Content, msg.AIMessage.Status, false)
	case MessageKindTool:
		if msg.ToolMessage == nil {
			return ""
		}
		return renderToolBubble(id, msg.ToolMessage.Name, msg.ToolMessage.Detail, false)
	case MessageKindError:
		if msg.ErrorMessage == nil {
			return ""
		}
		return renderErrorBubble(id, msg.ErrorMessage.Content, false)
	default:
		return ""
	}
}

func renderUserBubble(domID, content string, oobReplace bool) string {
	oob := ""
	if oobReplace {
		oob = ` hx-swap-oob="true"`
	}
	return fmt.Sprintf(
		`<div id="%s"%s class="chat chat-end mb-2"><div class="chat-bubble chat-bubble-primary whitespace-pre-wrap">%s</div></div>`,
		html.EscapeString(domID),
		oob,
		html.EscapeString(content),
	)
}

func renderAIBubble(domID, content, status string, oobReplace bool) string {
	oob := ""
	if oobReplace {
		oob = ` hx-swap-oob="true"`
	}
	extra := ""
	if status == AIStatusStreaming || status == AIStatusPending {
		extra = ` <span class="loading loading-dots loading-xs align-middle"></span>`
	}
	return fmt.Sprintf(
		`<div id="%s"%s class="chat chat-start mb-2"><div class="chat-bubble bg-base-200 text-base-content whitespace-pre-wrap">%s%s</div></div>`,
		html.EscapeString(domID),
		oob,
		html.EscapeString(content),
		extra,
	)
}

func renderToolBubble(domID, name, detail string, oobReplace bool) string {
	oob := ""
	if oobReplace {
		oob = ` hx-swap-oob="true"`
	}
	return fmt.Sprintf(
		`<div id="%s"%s class="alert alert-info mb-2 text-sm"><div class="font-mono text-xs">tool: %s</div><pre class="whitespace-pre-wrap mt-1">%s</pre></div>`,
		html.EscapeString(domID),
		oob,
		html.EscapeString(name),
		html.EscapeString(detail),
	)
}

func renderErrorBubble(domID, content string, oobReplace bool) string {
	oob := ""
	if oobReplace {
		oob = ` hx-swap-oob="true"`
	}
	return fmt.Sprintf(
		`<div id="%s"%s class="alert alert-error mb-2 text-sm whitespace-pre-wrap">%s</div>`,
		html.EscapeString(domID),
		oob,
		html.EscapeString(content),
	)
}

// OOBAppendTranscript wraps HTML to append into #sqlagent-transcript.
func OOBAppendTranscript(innerHTML string) string {
	var b strings.Builder
	b.WriteString(`<div hx-swap-oob="beforeend:#sqlagent-transcript">`)
	b.WriteString(innerHTML)
	b.WriteString(`</div>`)
	return b.String()
}

// OOBReplaceMessage replaces the bubble by id (for streaming updates).
func OOBReplaceMessage(msg ConversationMessage) string {
	if isRegistrySchemaToolMessage(&msg) {
		return ""
	}
	if msg.Kind == MessageKindUser && msg.UserMessage != nil && isRegistrySchemaBootstrapUserContent(msg.UserMessage.Content) {
		return ""
	}
	id := fmt.Sprintf("sqlagent-msg-%d", msg.ID)
	switch msg.Kind {
	case MessageKindUser:
		if msg.UserMessage == nil {
			return ""
		}
		return renderUserBubble(id, msg.UserMessage.Content, true)
	case MessageKindAI:
		if msg.AIMessage == nil {
			return ""
		}
		return renderAIBubble(id, msg.AIMessage.Content, msg.AIMessage.Status, true)
	case MessageKindTool:
		if msg.ToolMessage == nil {
			return ""
		}
		return renderToolBubble(id, msg.ToolMessage.Name, msg.ToolMessage.Detail, true)
	case MessageKindError:
		if msg.ErrorMessage == nil {
			return ""
		}
		return renderErrorBubble(id, msg.ErrorMessage.Content, true)
	default:
		return ""
	}
}

// LoadMessagesForConversation loads ordered messages with payloads for transcript rendering.
func LoadMessagesForConversation(db *gorm.DB, conversationID uint) ([]ConversationMessage, error) {
	chain := gorm.G[ConversationMessage](db).Where("conversation_id = ?", conversationID).
		Preload("UserMessage", nil).
		Preload("AIMessage", nil).
		Preload("ToolMessage", nil).
		Preload("ErrorMessage", nil).
		Order("sort_order ASC, id ASC")
	return chain.Find(context.Background())
}
