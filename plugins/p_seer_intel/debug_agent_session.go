package p_seer_intel

import (
	"encoding/json"
	"os"
	"time"
)

// #region agent log
const agentDebugLogPath = "/home/sandy/source_repos/lago/.cursor/debug-58e3a7.log"

// AgentDebugSessionLog appends one NDJSON line for Cursor debug mode. Do not pass secrets or PII.
func AgentDebugSessionLog(hypothesisID, location, message string, data map[string]any) {
	rec := map[string]any{
		"sessionId":    "58e3a7",
		"hypothesisId": hypothesisID,
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return
	}
	f, err := os.OpenFile(agentDebugLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	_, _ = f.Write(append(b, '\n'))
	_ = f.Close()
}

// #endregion
