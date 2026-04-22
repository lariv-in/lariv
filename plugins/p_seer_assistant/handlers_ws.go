package p_seer_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_users"
	"golang.org/x/net/websocket"
	"gorm.io/gorm"
)

func websocketUpgradeHandler() http.Handler {
	srv := websocket.Server{
		Handshake: func(_ *websocket.Config, _ *http.Request) error {
			return nil
		},
		Handler: assistantWebSocketConn,
	}
	return srv
}

func assistantWebSocketConn(ws *websocket.Conn) {
	req := ws.Request()
	if req == nil {
		return
	}
	ctx := req.Context()
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("p_seer_assistant: ws missing db", "error", err)
		return
	}
	buf := make([]byte, 64<<10)
	for {
		n, rerr := ws.Read(buf)
		if rerr != nil {
			return
		}
		if n <= 0 {
			continue
		}
		if err := handleOneWSClientPayload(ctx, db, ws, buf[:n]); err != nil {
			slog.Error("p_seer_assistant: ws handler", "error", err)
			_ = writeWSHTML(ws, errorOOB(err))
		}
	}
}

func handleOneWSClientPayload(ctx context.Context, db *gorm.DB, ws *websocket.Conn, payload []byte) error {
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	msg, _ := m["message"].(string)
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return fmt.Errorf("empty message")
	}
	sid, ok := parseSessionID(m["session_id"])
	if !ok {
		sid = 0
	}

	var userID *uint
	if u, ok := ctx.Value("$user").(p_users.User); ok {
		uid := u.ID
		userID = &uid
	}

	var session SeerAssistantSession
	if sid == 0 {
		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			s, err := CreateSession(ctx, tx, userID)
			if err != nil {
				return err
			}
			session = s
			if _, err := AppendUserMessage(ctx, tx, session.ID, msg); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		hidden := fmt.Sprintf(
			`<input id="seer_assistant_session_id" name="session_id" type="hidden" value="%d" hx-swap-oob="true"/>`,
			session.ID,
		)
		if err := writeWSHTML(ws, hidden); err != nil {
			return err
		}
		sid = session.ID
		ub := fmt.Sprintf(
			`<div id="seer_assistant_transcript" hx-swap-oob="beforeend"><div class="chat chat-end mb-2"><div class="chat-header text-xs opacity-70">You</div><div class="chat-bubble chat-bubble-primary whitespace-pre-wrap">%s</div></div></div>`,
			html.EscapeString(msg),
		)
		if err := writeWSHTML(ws, ub); err != nil {
			return err
		}
	} else {
		if err := db.WithContext(ctx).First(&session, sid).Error; err != nil {
			return fmt.Errorf("session not found: %w", err)
		}
		if session.UserID != nil && userID != nil && *session.UserID != *userID {
			return fmt.Errorf("session belongs to another user")
		}
		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			_, err := AppendUserMessage(ctx, tx, sid, msg)
			return err
		}); err != nil {
			return err
		}
		userBubble := fmt.Sprintf(
			`<div id="seer_assistant_transcript" hx-swap-oob="beforeend"><div class="chat chat-end mb-2"><div class="chat-header text-xs opacity-70">You</div><div class="chat-bubble chat-bubble-primary whitespace-pre-wrap">%s</div></div></div>`,
			html.EscapeString(msg),
		)
		if err := writeWSHTML(ws, userBubble); err != nil {
			return err
		}
	}

	return RunAssistantAfterUserMessage(ctx, db, ws, sid)
}

func parseSessionID(v any) (uint, bool) {
	switch x := v.(type) {
	case float64:
		if x <= 0 {
			return 0, false
		}
		return uint(x), true
	case string:
		s := strings.TrimSpace(x)
		if s == "" || s == "0" {
			return 0, false
		}
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil || n == 0 {
			return 0, false
		}
		return uint(n), true
	default:
		return 0, false
	}
}
