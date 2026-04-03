package sqlagent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"golang.org/x/net/websocket"
	"gorm.io/gorm"
)

// wsPayload is the JSON shape sent by htmx-ext-ws on ws-send (form fields + HEADERS).
type wsPayload map[string]json.RawMessage

func wsHTTPHandler() http.Handler {
	s := websocket.Server{
		Handler: chatSocketHandler,
		// Skip default Origin check from websocket.Handler; AuthenticationMiddleware already ran.
		Handshake: func(_ *websocket.Config, _ *http.Request) error { return nil },
	}
	return p_users.AuthenticationMiddleware{}.Next(views.View{}, http.HandlerFunc(s.ServeHTTP))
}

func chatSocketHandler(conn *websocket.Conn) {
	req := conn.Request()
	if req == nil {
		slog.Error("sqlagent: websocket request is nil")
		return
	}
	ctx := req.Context()
	u, ok := ctx.Value("$user").(p_users.User)
	if !ok {
		slog.Error("sqlagent: websocket missing $user")
		return
	}
	convIDStr := req.PathValue("conversation_id")
	convID64, err := strconv.ParseUint(convIDStr, 10, 64)
	if err != nil || convID64 == 0 {
		slog.Error("sqlagent: invalid conversation_id in websocket", "raw", convIDStr, "error", err)
		return
	}
	conversationID := uint(convID64)

	db := ctx.Value("$db").(*gorm.DB)
	conv, err := gorm.G[Conversation](db).Where("id = ? AND created_by_id = ?", conversationID, u.ID).First(ctx)
	if err != nil {
		slog.Error("sqlagent: websocket conversation not found or denied", "error", err, "conversation_id", conversationID)
		return
	}

	for {
		var raw string
		if err := websocket.Message.Receive(conn, &raw); err != nil {
			if !strings.Contains(err.Error(), "EOF") && !strings.Contains(err.Error(), "closed") {
				slog.Error("sqlagent: websocket receive", "error", err)
			}
			return
		}
		content := parseWSContent(raw)
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}
		if err := handleChatMessage(conn, db, &conv, u.ID, content); err != nil {
			slog.Error("sqlagent: handle chat message", "error", err)
			errID := fmt.Sprintf("sqlagent-ws-err-%d", time.Now().UnixNano())
			_ = websocket.Message.Send(conn, OOBAppendTranscript(renderErrorBubble(errID, err.Error(), false)))
		}
	}
}

func parseWSContent(raw string) string {
	var p wsPayload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return raw
	}
	if v, ok := p["content"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			return s
		}
	}
	return ""
}

func nextSortOrder(db *gorm.DB, conversationID uint) (int, error) {
	last, err := gorm.G[ConversationMessage](db).Where("conversation_id = ?", conversationID).Order("sort_order DESC").First(context.Background())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return last.SortOrder + 1, nil
}

func handleChatMessage(conn *websocket.Conn, db *gorm.DB, conv *Conversation, userID uint, content string) error {
	if conv.CreatedByID != userID {
		return errors.New("forbidden")
	}

	var userMsg ConversationMessage
	var aiMsg ConversationMessage

	err := db.Transaction(func(tx *gorm.DB) error {
		sortU, err := nextSortOrder(tx, conv.ID)
		if err != nil {
			return err
		}
		userMsg = ConversationMessage{
			ConversationID: conv.ID,
			SortOrder:      sortU,
			Kind:           MessageKindUser,
		}
		if err := gorm.G[ConversationMessage](tx).Create(context.Background(), &userMsg); err != nil {
			return err
		}
		um := UserMessage{
			ConversationMessageID: userMsg.ID,
			Content:               content,
		}
		if err := gorm.G[UserMessage](tx).Create(context.Background(), &um); err != nil {
			return err
		}
		userMsg.UserMessage = &um

		sortA := sortU + 1
		aiMsg = ConversationMessage{
			ConversationID: conv.ID,
			SortOrder:      sortA,
			Kind:           MessageKindAI,
		}
		if err := gorm.G[ConversationMessage](tx).Create(context.Background(), &aiMsg); err != nil {
			return err
		}
		am := AIMessage{
			ConversationMessageID: aiMsg.ID,
			Content:               "",
			Status:                AIStatusStreaming,
		}
		if err := gorm.G[AIMessage](tx).Create(context.Background(), &am); err != nil {
			return err
		}
		aiMsg.AIMessage = &am
		return nil
	})
	if err != nil {
		return err
	}

	maybeUpdateConversationTitle(db, conv.ID, content)

	if err := db.Model(&Conversation{}).Where("id = ?", conv.ID).Update("updated_at", time.Now()).Error; err != nil {
		slog.Error("sqlagent: touch conversation", "error", err)
	}

	if err := websocket.Message.Send(conn, OOBAppendTranscript(RenderMessageBubble(userMsg))); err != nil {
		return err
	}
	if err := websocket.Message.Send(conn, OOBAppendTranscript(RenderMessageBubble(aiMsg))); err != nil {
		return err
	}

	// Simulated streamed assistant response (no LLM in this phase).
	full := "This is a simulated assistant reply. Your message was received over the WebSocket. SQL tooling is not wired yet."
	return streamAssistantReply(conn, db, &aiMsg, full)
}

func streamAssistantReply(conn *websocket.Conn, db *gorm.DB, aiEnvelope *ConversationMessage, full string) error {
	words := strings.Fields(full)
	if len(words) == 0 {
		words = []string{""}
	}
	var b strings.Builder
	for i, w := range words {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(w)
		chunk := b.String()
		if err := db.Model(&AIMessage{}).Where("conversation_message_id = ?", aiEnvelope.ID).
			Updates(map[string]any{"content": chunk}).Error; err != nil {
			slog.Error("sqlagent: stream update ai message", "error", err)
			return err
		}
		aiEnvelope.AIMessage.Content = chunk
		aiEnvelope.AIMessage.Status = AIStatusStreaming
		if err := websocket.Message.Send(conn, OOBReplaceMessage(*aiEnvelope)); err != nil {
			return err
		}
		time.Sleep(40 * time.Millisecond)
	}
	if err := db.Model(&AIMessage{}).Where("conversation_message_id = ?", aiEnvelope.ID).
		Updates(map[string]any{"status": AIStatusComplete}).Error; err != nil {
		return err
	}
	aiEnvelope.AIMessage.Status = AIStatusComplete
	return websocket.Message.Send(conn, OOBReplaceMessage(*aiEnvelope))
}

// Sidebar title preview: update conversation title from first user message if still default.
func maybeUpdateConversationTitle(db *gorm.DB, convID uint, titleHint string) {
	titleHint = strings.TrimSpace(titleHint)
	if titleHint == "" {
		return
	}
	c, err := gorm.G[Conversation](db).Where("id = ?", convID).First(context.Background())
	if err != nil {
		slog.Error("sqlagent: load conversation for title", "error", err)
		return
	}
	if c.Title != "" && c.Title != "New conversation" {
		return
	}
	runes := []rune(titleHint)
	if len(runes) > 80 {
		titleHint = string(runes[:80]) + "…"
	}
	if err := db.Model(&Conversation{}).Where("id = ?", convID).Update("title", titleHint).Error; err != nil {
		slog.Error("sqlagent: update conversation title", "error", err)
	}
}
