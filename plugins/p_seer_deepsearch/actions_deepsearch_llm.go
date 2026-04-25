package p_seer_deepsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/lariv-in/lago/plugins/p_google_genai"
	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

const deepSearchReportMaxToolRounds = 22

func deepSearchExpandMaxOutputTokens() int32 {
	n := DeepSearchAppConfig.ExpandMaxOutputTokens
	if n <= 0 {
		n = defaultDeepSearchExpandMaxOutputTokens
	}
	if n > maxDeepSearchExpandMaxOutputTokens {
		n = maxDeepSearchExpandMaxOutputTokens
	}
	return int32(n)
}

func deepSearchReportMaxOutputTokens() int32 {
	n := DeepSearchAppConfig.ReportMaxOutputTokens
	if n <= 0 {
		n = defaultDeepSearchReportMaxOutputTokens
	}
	if n > maxDeepSearchReportMaxOutputTokens {
		n = maxDeepSearchReportMaxOutputTokens
	}
	return int32(n)
}

func deepsearchModel() string {
	if DeepSearchAppConfig == nil {
		return defaultDeepSearchLlmModel
	}
	m := strings.TrimSpace(DeepSearchAppConfig.LlmModel)
	if m == "" {
		return defaultDeepSearchLlmModel
	}
	return m
}

func deepsearchGenerateContentJSON[T any](ctx context.Context, system, user string, maxOut int32, temp *float32, out *T) (raw string, err error) {
	client, err := p_google_genai.NewClient(ctx)
	if err != nil {
		return "", err
	}
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(system, genai.RoleUser),
		ResponseMIMEType:  "application/json",
		ResponseSchema:    p_google_genai.NewSchema[T](),
		MaxOutputTokens:   maxOut,
	}
	if temp != nil {
		cfg.Temperature = temp
	}
	resp, err := client.Models.GenerateContent(ctx, deepsearchModel(), []*genai.Content{genai.NewContentFromText(user, genai.RoleUser)}, cfg)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("p_seer_deepsearch: nil generate response")
	}
	raw = strings.TrimSpace(resp.Text())
	if raw == "" {
		return "", fmt.Errorf("p_seer_deepsearch: empty model response")
	}
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		return raw, fmt.Errorf("p_seer_deepsearch: json decode: %w", err)
	}
	return raw, nil
}

// expandDeepSearchQueries asks Gemini ([p_google_genai]) for distinct web search query strings (JSON array).
func expandDeepSearchQueries(ctx context.Context, userQuery string) ([]string, error) {
	userQuery = strings.TrimSpace(userQuery)
	if userQuery == "" {
		return nil, fmt.Errorf("p_seer_deepsearch: empty user query")
	}
	// #region agent log
	p_seer_intel.AgentDebugSessionLog("H4", "actions_deepsearch_llm.go:expandDeepSearchQueries", "expand_start", map[string]any{
		"queryLen":  utf8.RuneCountInString(userQuery),
		"maxOutTok": deepSearchExpandMaxOutputTokens(),
	})
	// #endregion
	sys := `You expand a user's research question into several distinct Google web search queries.
Return ONLY a JSON array of strings (no markdown, no keys, no commentary).
Use search operators when helpful: site:, filetype:, exact phrases in double quotes, OR groups.
Prefer high-recall, diverse angles. At most 8 strings.`

	var arr []string
	expandTemp := float32(0.35)
	_, err := deepsearchGenerateContentJSON(ctx, sys, userQuery, deepSearchExpandMaxOutputTokens(), &expandTemp, &arr)
	if err != nil {
		slog.Error("p_seer_deepsearch: expand queries", "error", err)
		return nil, fmt.Errorf("p_seer_deepsearch: expand queries: %w", err)
	}
	var out []string
	seen := make(map[string]struct{}, len(arr))
	for _, s := range arr {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
		if len(out) >= maxGeneratedSearchQueries {
			break
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("p_seer_deepsearch: no queries after parse")
	}
	return out, nil
}

type deepSearchToolDecision struct {
	Action  string `json:"action"`
	Query   string `json:"query,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	IntelID uint   `json:"intel_id,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

const deepSearchReportLogToolPayloadMaxRunes = 14000

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max]) + "\n…(truncated)"
}

func runDeepSearchReport(ctx context.Context, db *gorm.DB, deepSearchID uint, userQuery string) (string, error) {
	user := strings.TrimSpace(userQuery)
	transcript := make([]string, 0, deepSearchReportMaxToolRounds*2)
	for round := 0; round < deepSearchReportMaxToolRounds; round++ {
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("p_seer_deepsearch: report stopped: %w", err)
		}
		decision, raw, err := nextDeepSearchToolDecision(ctx, user, transcript, round)
		if err != nil {
			slog.Error("p_seer_deepsearch: report decision", "error", err, "round", round)
			return "", fmt.Errorf("p_seer_deepsearch: report decision: %w", err)
		}
		if deepSearchID != 0 {
			appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindReportLlm,
				fmt.Sprintf("round=%d decision=%s", round, truncateRunes(raw, deepSearchReportLogToolPayloadMaxRunes)))
		}
		action := strings.ToLower(strings.TrimSpace(decision.Action))
		if action == "" {
			return "", fmt.Errorf("p_seer_deepsearch: empty tool action (round %d)", round)
		}
		if action == "final_report" {
			return generateFinalDeepSearchReport(ctx, user, transcript)
		}
		payload, terr := dispatchDeepSearchDecision(ctx, db, decision)
		if terr != nil {
			slog.Warn("p_seer_deepsearch: tool error", "tool", action, "error", terr)
			transcript = append(transcript,
				fmt.Sprintf("Round %d decision: %s", round, raw),
				fmt.Sprintf("Round %d tool error: %s", round, terr.Error()),
			)
			if deepSearchID != 0 {
				appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindReportLlm,
					fmt.Sprintf("round=%d tool=%s ERROR: %s", round, action, terr.Error()))
			}
			continue
		}
		payloadJSON, _ := json.Marshal(payload)
		transcript = append(transcript,
			fmt.Sprintf("Round %d decision: %s", round, raw),
			fmt.Sprintf("Round %d tool result: %s", round, string(payloadJSON)),
		)
		if deepSearchID != 0 {
			appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindReportLlm,
				fmt.Sprintf("round=%d tool=%s result=%s", round, action, truncateRunes(string(payloadJSON), deepSearchReportLogToolPayloadMaxRunes)))
		}
	}
	return "", fmt.Errorf("p_seer_deepsearch: report exceeded tool rounds")
}

func nextDeepSearchToolDecision(ctx context.Context, userQuery string, transcript []string, round int) (deepSearchToolDecision, string, error) {
	var out deepSearchToolDecision
	sys := `You are a research assistant controlling two tools over an Intel database.

Return ONLY JSON with one action at a time:
- {"action":"intel_vector_search","query":"...", "limit":8, "reason":"..."}
- {"action":"intel_get_content","intel_id":123, "reason":"..."}
- {"action":"final_report","reason":"..."}

Rules:
- Prefer intel_vector_search before final_report.
- After relevant vector hits, fetch full content for promising ids with intel_get_content before final_report.
- Do not invent tool names or extra top-level fields.`
	var prompt strings.Builder
	prompt.WriteString("User question:\n")
	prompt.WriteString(userQuery)
	prompt.WriteString("\n\nTool transcript so far:\n")
	if len(transcript) == 0 {
		prompt.WriteString("(none yet)")
	} else {
		prompt.WriteString(strings.Join(transcript, "\n\n"))
	}
	prompt.WriteString(fmt.Sprintf("\n\nChoose next action for round %d.", round))
	raw, err := deepsearchGenerateContentJSON(ctx, sys, prompt.String(), 512, nil, &out)
	return out, raw, err
}

func generateFinalDeepSearchReport(ctx context.Context, userQuery string, transcript []string) (string, error) {
	sys := `You are a research assistant. The user asked a question; relevant web pages were scraped and summarized into an Intel database.

Your final answer must be a long, detailed markdown report. Aim for depth and length: multiple sections with ## and ### headings, rich paragraphs, and synthesis across sources. Ground every substantive claim in the provided Intel tool transcript.

Structure suggestions (adapt to the question): title line (# heading), executive overview, thematic sections with analysis, a "Sources / evidence" subsection, gaps or uncertainties, and optional brief conclusion. Cite Intel using markdown links only: [short label](intel://<intel_id>) where <intel_id> is the numeric id from the transcript. Do not use plain-text forms like "Intel #42".

If evidence is thin, say so explicitly, but still write a careful report explaining what was and was not found.

Output markdown only.`
	var prompt strings.Builder
	prompt.WriteString("User question:\n")
	prompt.WriteString(strings.TrimSpace(userQuery))
	prompt.WriteString("\n\nIntel tool transcript:\n")
	if len(transcript) == 0 {
		prompt.WriteString("(no tool transcript)")
	} else {
		prompt.WriteString(strings.Join(transcript, "\n\n"))
	}
	reportTemp := float32(0.25)
	client, err := p_google_genai.NewClient(ctx)
	if err != nil {
		return "", err
	}
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(sys, genai.RoleUser),
		Temperature:       &reportTemp,
		MaxOutputTokens:   deepSearchReportMaxOutputTokens(),
	}
	resp, err := client.Models.GenerateContent(ctx, deepsearchModel(), []*genai.Content{genai.NewContentFromText(prompt.String(), genai.RoleUser)}, cfg)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("p_seer_deepsearch: nil generate response")
	}
	return strings.TrimSpace(resp.Text()), nil
}

func dispatchDeepSearchDecision(ctx context.Context, db *gorm.DB, decision deepSearchToolDecision) (any, error) {
	switch strings.ToLower(strings.TrimSpace(decision.Action)) {
	case "intel_vector_search":
		q := strings.TrimSpace(decision.Query)
		if q == "" {
			return nil, fmt.Errorf("missing query")
		}
		limit := decision.Limit
		if limit == 0 {
			limit = 8
		}
		if limit < 1 {
			limit = 1
		}
		if limit > 25 {
			limit = 25
		}
		hits, err := p_seer_intel.SearchIntelBySimilarity(ctx, db, q, limit)
		if err != nil {
			return nil, err
		}
		type row struct {
			ID      uint   `json:"id"`
			Title   string `json:"title"`
			Summary string `json:"summary"`
		}
		var rows []row
		for _, h := range hits {
			rows = append(rows, row{
				ID:      h.ID,
				Title:   h.Title,
				Summary: h.Summary,
			})
		}
		return rows, nil
	case "intel_get_content":
		if decision.IntelID == 0 {
			return nil, fmt.Errorf("missing intel_id")
		}
		doc, err := p_seer_intel.IntelDocumentForTool(ctx, db, decision.IntelID)
		if err != nil {
			return nil, err
		}
		return doc, nil
	default:
		return nil, fmt.Errorf("unknown tool %q", decision.Action)
	}
}
