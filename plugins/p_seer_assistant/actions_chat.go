package p_seer_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AllocateTranscriptOrdinal reserves the next ordinal for a session row (transactional).
func AllocateTranscriptOrdinal(tx *gorm.DB, sessionID uint) (uint, error) {
	var s SeerAssistantSession
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", sessionID).First(&s).Error; err != nil {
		return 0, err
	}
	next := s.NextTranscriptOrdinal + 1
	if err := tx.Model(&SeerAssistantSession{}).Where("id = ?", sessionID).
		Update("next_transcript_ordinal", next).Error; err != nil {
		return 0, err
	}
	return next, nil
}

// CreateSession inserts a new [SeerAssistantSession].
func CreateSession(ctx context.Context, db *gorm.DB, userID *uint) (SeerAssistantSession, error) {
	s := SeerAssistantSession{UserID: userID}
	if err := db.WithContext(ctx).Create(&s).Error; err != nil {
		return SeerAssistantSession{}, err
	}
	return s, nil
}

// AppendUserMessage inserts a user row at the next ordinal.
func AppendUserMessage(ctx context.Context, tx *gorm.DB, sessionID uint, body string) (SeerAssistantUserMessage, error) {
	ord, err := AllocateTranscriptOrdinal(tx, sessionID)
	if err != nil {
		return SeerAssistantUserMessage{}, err
	}
	m := SeerAssistantUserMessage{SessionID: sessionID, Ordinal: ord, Body: body}
	if err := tx.WithContext(ctx).Create(&m).Error; err != nil {
		return SeerAssistantUserMessage{}, err
	}
	return m, nil
}

// AppendAssistantMessage stores the final assistant text for one turn.
func AppendAssistantMessage(ctx context.Context, tx *gorm.DB, sessionID uint, body string) (SeerAssistantAssistantMessage, error) {
	ord, err := AllocateTranscriptOrdinal(tx, sessionID)
	if err != nil {
		return SeerAssistantAssistantMessage{}, err
	}
	m := SeerAssistantAssistantMessage{SessionID: sessionID, Ordinal: ord, Body: body}
	if err := tx.WithContext(ctx).Create(&m).Error; err != nil {
		return SeerAssistantAssistantMessage{}, err
	}
	return m, nil
}

// AppendToolCall persists a tool invocation row.
func AppendToolCall(ctx context.Context, tx *gorm.DB, sessionID uint, name string, arguments any) (SeerAssistantToolCall, error) {
	ord, err := AllocateTranscriptOrdinal(tx, sessionID)
	if err != nil {
		return SeerAssistantToolCall{}, err
	}
	b, err := json.Marshal(arguments)
	if err != nil {
		return SeerAssistantToolCall{}, err
	}
	m := SeerAssistantToolCall{SessionID: sessionID, Ordinal: ord, Name: name, Arguments: b}
	if err := tx.WithContext(ctx).Create(&m).Error; err != nil {
		return SeerAssistantToolCall{}, err
	}
	return m, nil
}

// AppendToolResult persists the tool output linked to a call.
func AppendToolResult(ctx context.Context, tx *gorm.DB, sessionID uint, toolCallID uint, result string, errText string) (SeerAssistantToolResult, error) {
	ord, err := AllocateTranscriptOrdinal(tx, sessionID)
	if err != nil {
		return SeerAssistantToolResult{}, err
	}
	m := SeerAssistantToolResult{
		ToolCallID: toolCallID,
		SessionID:  sessionID,
		Ordinal:    ord,
		Result:     result,
		Error:      errText,
	}
	if err := tx.WithContext(ctx).Create(&m).Error; err != nil {
		return SeerAssistantToolResult{}, err
	}
	return m, nil
}

// BuildChatTurns reconstructs model input from persisted per-kind tables.
func BuildChatTurns(ctx context.Context, db *gorm.DB, sessionID uint) ([]AssistantChatTurn, error) {
	var users []SeerAssistantUserMessage
	if err := db.WithContext(ctx).Where("session_id = ?", sessionID).Order("ordinal ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	var assts []SeerAssistantAssistantMessage
	if err := db.WithContext(ctx).Where("session_id = ?", sessionID).Order("ordinal ASC").Find(&assts).Error; err != nil {
		return nil, err
	}
	var calls []SeerAssistantToolCall
	if err := db.WithContext(ctx).Where("session_id = ?", sessionID).Order("ordinal ASC").Find(&calls).Error; err != nil {
		return nil, err
	}
	var results []SeerAssistantToolResult
	if err := db.WithContext(ctx).Where("session_id = ?", sessionID).Order("ordinal ASC").Find(&results).Error; err != nil {
		return nil, err
	}

	type tagged struct {
		ord uint
		typ string
		msg AssistantChatTurn
	}
	var slots []tagged
	for _, u := range users {
		slots = append(slots, tagged{ord: u.Ordinal, typ: "user", msg: AssistantChatTurn{Role: "user", Content: u.Body}})
	}
	for _, a := range assts {
		slots = append(slots, tagged{ord: a.Ordinal, typ: "assistant", msg: AssistantChatTurn{Role: "assistant", Content: a.Body}})
	}
	resByCall := make(map[uint]SeerAssistantToolResult, len(results))
	for _, r := range results {
		resByCall[r.ToolCallID] = r
	}
	for _, c := range calls {
		args := string(c.Arguments)
		res, ok := resByCall[c.ID]
		body := fmt.Sprintf("[tool %s] args: %s\n", c.Name, args)
		if ok {
			if res.Error != "" {
				body += fmt.Sprintf("[error] %s", res.Error)
			} else {
				body += fmt.Sprintf("[result] %s", res.Result)
			}
		} else {
			body += "[result] <pending>"
		}
		slots = append(slots, tagged{ord: c.Ordinal, typ: "tool", msg: AssistantChatTurn{Role: "user", Content: body}})
	}

	sort.Slice(slots, func(i, j int) bool {
		if slots[i].ord != slots[j].ord {
			return slots[i].ord < slots[j].ord
		}
		// deterministic tiebreak: user, assistant, tool
		prio := map[string]int{"user": 0, "assistant": 1, "tool": 2}
		return prio[slots[i].typ] < prio[slots[j].typ]
	})

	out := make([]AssistantChatTurn, 0, len(slots))
	for _, s := range slots {
		out = append(out, s.msg)
	}
	return out, nil
}
