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
		// Skip default Origin check from websocket.Handler; AuthenticationLayer already ran.
		Handshake: func(_ *websocket.Config, _ *http.Request) error { return nil },
	}
	return p_users.AuthenticationLayer{}.Next(views.View{}, http.HandlerFunc(s.ServeHTTP))
}

func chatSocketHandler(conn *websocket.Conn) {
	req := conn.Request()
	if req == nil {
		logError("sqlagent: websocket request is nil", errors.New("nil request"))
		return
	}
	ctx := req.Context()
	u, ok := ctx.Value("$user").(p_users.User)
	if !ok {
		logError("sqlagent: websocket missing $user", errors.New("no $user in context"))
		return
	}
	convIDStr := req.PathValue("conversation_id")
	convID64, err := strconv.ParseUint(convIDStr, 10, 64)
	if err != nil || convID64 == 0 {
		if err == nil {
			err = errors.New("conversation_id is zero or invalid")
		}
		logError("sqlagent: invalid conversation_id in websocket", err, "raw", convIDStr)
		return
	}
	conversationID := uint(convID64)

	db := ctx.Value("$db").(*gorm.DB)
	conv, err := gorm.G[Conversation](db).Where("id = ? AND created_by_id = ?", conversationID, u.ID).First(ctx)
	if err != nil {
		logError("sqlagent: websocket conversation not found or denied", err, "conversation_id", conversationID)
		return
	}

	for {
		var raw string
		if err := websocket.Message.Receive(conn, &raw); err != nil {
			if !strings.Contains(err.Error(), "EOF") && !strings.Contains(err.Error(), "closed") {
				logError("sqlagent: websocket receive", err)
			}
			return
		}
		content := parseWSContent(raw)
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}
		if err := handleChatMessage(conn, db, &conv, u.ID, content); err != nil {
			logError("sqlagent: websocket chat turn", err, "conversation_id", conversationID)
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
		logError("sqlagent: next sort order query", err, "conversation_id", conversationID)
		return 0, err
	}
	return last.SortOrder + 1, nil
}

func handleChatMessage(conn *websocket.Conn, db *gorm.DB, conv *Conversation, userID uint, content string) error {
	if conv.CreatedByID != userID {
		err := errors.New("forbidden")
		logError("sqlagent: forbidden chat message", err, "conversation_id", conv.ID, "user_id", userID)
		return err
	}

	var userMsg ConversationMessage
	var aiMsg ConversationMessage

	err := db.Transaction(func(tx *gorm.DB) error {
		sortU, err := nextSortOrder(tx, conv.ID)
		if err != nil {
			logError("sqlagent: next sort order", err, "conversation_id", conv.ID)
			return err
		}
		userMsg = ConversationMessage{
			ConversationID: conv.ID,
			SortOrder:      sortU,
			Kind:           MessageKindUser,
		}
		if err := gorm.G[ConversationMessage](tx).Create(context.Background(), &userMsg); err != nil {
			logError("sqlagent: create user message envelope", err, "conversation_id", conv.ID)
			return err
		}
		um := UserMessage{
			ConversationMessageID: userMsg.ID,
			Content:               content,
		}
		if err := gorm.G[UserMessage](tx).Create(context.Background(), &um); err != nil {
			logError("sqlagent: create user message body", err, "conversation_id", conv.ID)
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
			logError("sqlagent: create ai message envelope", err, "conversation_id", conv.ID)
			return err
		}
		am := AIMessage{
			ConversationMessageID: aiMsg.ID,
			Content:               "",
			Status:                AIStatusStreaming,
		}
		if err := gorm.G[AIMessage](tx).Create(context.Background(), &am); err != nil {
			logError("sqlagent: create ai message body", err, "conversation_id", conv.ID)
			return err
		}
		aiMsg.AIMessage = &am
		return nil
	})
	if err != nil {
		logError("sqlagent: persist chat messages transaction", err, "conversation_id", conv.ID)
		return err
	}

	maybeUpdateConversationTitle(db, conv.ID, content)

	if err := db.Model(&Conversation{}).Where("id = ?", conv.ID).Update("updated_at", time.Now()).Error; err != nil {
		logError("sqlagent: touch conversation", err, "conversation_id", conv.ID)
	}

	if err := websocket.Message.Send(conn, OOBAppendTranscript(RenderMessageBubble(userMsg))); err != nil {
		logError("sqlagent: websocket send user bubble", err, "conversation_id", conv.ID)
		return err
	}
	if err := websocket.Message.Send(conn, OOBAppendTranscript(RenderMessageBubble(aiMsg))); err != nil {
		logError("sqlagent: websocket send ai placeholder bubble", err, "conversation_id", conv.ID)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	rt, err := loadADK(ctx)
	if err != nil {
		logError("sqlagent: load ADK", err, "conversation_id", conv.ID)
		slog.Info("sqlagent: ADK unavailable, using simulated reply", "error", err)
		full := "This is a simulated assistant reply. Set GOOGLE_API_KEY or GEMINI_API_KEY to use Gemini via ADK (the live assistant can run raw SQL with the execute_sql tool). Your message was received over the WebSocket."
		if err := streamAssistantReply(conn, db, &aiMsg, full); err != nil {
			logError("sqlagent: simulated stream reply", err, "conversation_id", conv.ID)
			return err
		}
		return nil
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		runCtx := ContextWithGormTx(ctx, tx)
		if err := seedADKSessionFromDB(runCtx, rt, tx, userID, conv.ID, userMsg.SortOrder); err != nil {
			return err
		}
		var lastText string
		err = ForEachADKReplyChunk(runCtx, rt, userID, conv.ID, content, func(text string) error {
			if text == "" {
				return nil
			}
			lastText = text
			if err := tx.Model(&AIMessage{}).Where("conversation_message_id = ?", aiMsg.ID).
				Updates(map[string]any{"content": text}).Error; err != nil {
				logError("sqlagent: stream update ai message content", err, "conversation_id", conv.ID, "ai_message_id", aiMsg.ID)
				return err
			}
			aiMsg.AIMessage.Content = text
			aiMsg.AIMessage.Status = AIStatusStreaming
			if sendErr := websocket.Message.Send(conn, OOBReplaceMessage(aiMsg)); sendErr != nil {
				logError("sqlagent: websocket OOB replace streaming ai bubble", sendErr, "conversation_id", conv.ID)
				return sendErr
			}
			return nil
		})
		if err != nil {
			return err
		}
		if lastText == "" {
			lastText = "(No response text from the model.)"
			if err := tx.Model(&AIMessage{}).Where("conversation_message_id = ?", aiMsg.ID).
				Updates(map[string]any{"content": lastText}).Error; err != nil {
				logError("sqlagent: set empty ai reply placeholder", err, "conversation_id", conv.ID)
				return err
			}
			aiMsg.AIMessage.Content = lastText
			if err := websocket.Message.Send(conn, OOBReplaceMessage(aiMsg)); err != nil {
				logError("sqlagent: websocket OOB replace empty ai reply", err, "conversation_id", conv.ID)
				return err
			}
		}
		if err := tx.Model(&AIMessage{}).Where("conversation_message_id = ?", aiMsg.ID).
			Updates(map[string]any{"status": AIStatusComplete}).Error; err != nil {
			logError("sqlagent: finalize ai message status", err, "conversation_id", conv.ID)
			return err
		}
		aiMsg.AIMessage.Status = AIStatusComplete
		if err := websocket.Message.Send(conn, OOBReplaceMessage(aiMsg)); err != nil {
			logError("sqlagent: websocket OOB replace final ai bubble", err, "conversation_id", conv.ID)
			return err
		}
		return nil
	})
	if err != nil {
		logError("sqlagent: ADK reply transaction", err, "conversation_id", conv.ID)
	}
	return err
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
			logError("sqlagent: simulated stream update ai message", err, "conversation_message_id", aiEnvelope.ID)
			return err
		}
		aiEnvelope.AIMessage.Content = chunk
		aiEnvelope.AIMessage.Status = AIStatusStreaming
		if err := websocket.Message.Send(conn, OOBReplaceMessage(*aiEnvelope)); err != nil {
			logError("sqlagent: simulated stream websocket send", err, "conversation_message_id", aiEnvelope.ID)
			return err
		}
		time.Sleep(40 * time.Millisecond)
	}
	if err := db.Model(&AIMessage{}).Where("conversation_message_id = ?", aiEnvelope.ID).
		Updates(map[string]any{"status": AIStatusComplete}).Error; err != nil {
		logError("sqlagent: simulated stream finalize ai status", err, "conversation_message_id", aiEnvelope.ID)
		return err
	}
	aiEnvelope.AIMessage.Status = AIStatusComplete
	if err := websocket.Message.Send(conn, OOBReplaceMessage(*aiEnvelope)); err != nil {
		logError("sqlagent: simulated stream final websocket send", err, "conversation_message_id", aiEnvelope.ID)
		return err
	}
	return nil
}

// Sidebar title preview: update conversation title from first user message if still default.
func maybeUpdateConversationTitle(db *gorm.DB, convID uint, titleHint string) {
	titleHint = strings.TrimSpace(titleHint)
	if titleHint == "" || isRegistrySchemaBootstrapUserContent(titleHint) {
		return
	}
	c, err := gorm.G[Conversation](db).Where("id = ?", convID).First(context.Background())
	if err != nil {
		logError("sqlagent: load conversation for title", err, "conversation_id", convID)
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
		logError("sqlagent: update conversation title", err, "conversation_id", convID)
	}
}
