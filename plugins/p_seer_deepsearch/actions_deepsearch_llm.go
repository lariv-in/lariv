package p_seer_deepsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"unicode/utf8"

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

func newGenAIClient(ctx context.Context) (*genai.Client, string, error) {
	key := strings.TrimSpace(p_seer_intel.IntelGenAI.APIKey)
	if key == "" {
		return nil, "", fmt.Errorf("p_seer_deepsearch: Plugins.p_seer_intel apiKey is empty")
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  key,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, "", err
	}
	model := strings.TrimSpace(p_seer_intel.IntelGenAI.LLMModel)
	if model == "" {
		model = "gemini-2.0-flash"
	}
	return client, model, nil
}

// expandDeepSearchQueries asks the LLM for distinct web search query strings (JSON array).
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
	client, model, err := newGenAIClient(ctx)
	if err != nil {
		return nil, err
	}

	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "string",
		},
		"minItems": 1,
		"maxItems": maxGeneratedSearchQueries,
	}

	sys := `You expand a user's research question into several distinct Google web search queries.
Return ONLY a JSON array of strings (no markdown, no keys, no commentary).
Use search operators when helpful: site:, filetype:, exact phrases in double quotes, OR groups.
Prefer high-recall, diverse angles. At most ` + strconv.Itoa(maxGeneratedSearchQueries) + ` strings.`

	cfg := &genai.GenerateContentConfig{
		SystemInstruction:  genai.NewContentFromText(sys, genai.RoleUser),
		ResponseMIMEType:   "application/json",
		ResponseJsonSchema: schema,
		Temperature:        genai.Ptr[float32](0.35),
		MaxOutputTokens:    deepSearchExpandMaxOutputTokens(),
	}

	resp, err := p_seer_intel.WithGenAIRetry(ctx, "deepsearch.expand_queries", func(ctx context.Context) (*genai.GenerateContentResponse, error) {
		return client.Models.GenerateContent(ctx, model, genai.Text(userQuery), cfg)
	})
	if err != nil {
		slog.Error("p_seer_deepsearch: expand queries", "error", err)
		return nil, fmt.Errorf("p_seer_deepsearch: expand queries: %w", err)
	}
	raw := strings.TrimSpace(resp.Text())
	if raw == "" {
		return nil, fmt.Errorf("p_seer_deepsearch: expand queries returned empty")
	}
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return nil, fmt.Errorf("p_seer_deepsearch: expand queries json: %w", err)
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

func reportTools() []*genai.Tool {
	querySchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{"type": "string", "description": "Natural language search over Intel embeddings."},
			"limit": map[string]any{"type": "integer", "description": "Max hits (default 8)."},
		},
		"required": []string{"query"},
	}
	getSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"intel_id": map[string]any{"type": "integer", "description": "Intel row primary key."},
		},
		"required": []string{"intel_id"},
	}
	return []*genai.Tool{{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:                 "intel_vector_search",
				Description:          "Find Intel rows similar to a text query (vector search). Returns ids, titles, summaries (best matches first). In the final markdown report, cite each Intel using [label](intel://<id>) with the returned id.",
				ParametersJsonSchema: querySchema,
			},
			{
				Name:                 "intel_get_content",
				Description:          "Load full Intel metadata and underlying source text for one Intel id. In the final markdown report, cite this Intel using [label](intel://<intel_id>).",
				ParametersJsonSchema: getSchema,
			},
		},
	}}
}

func argString(m map[string]any, key string) (string, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return "", false
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t), true
	default:
		return strings.TrimSpace(fmt.Sprint(t)), true
	}
}

func argUint(m map[string]any, key string) (uint, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		if t < 0 || t > float64(^uint(0)>>1) {
			return 0, false
		}
		return uint(t), true
	case int:
		if t < 0 {
			return 0, false
		}
		return uint(t), true
	case int64:
		if t < 0 {
			return 0, false
		}
		return uint(t), true
	default:
		n, err := strconv.ParseUint(strings.TrimSpace(fmt.Sprint(v)), 10, 32)
		if err != nil {
			return 0, false
		}
		return uint(n), true
	}
}

func argIntDefault(m map[string]any, key string, def int) int {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		n, err := strconv.Atoi(strings.TrimSpace(fmt.Sprint(v)))
		if err != nil {
			return def
		}
		return n
	}
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

func summarizeModelContentForLog(c *genai.Content) string {
	if c == nil {
		return ""
	}
	var b strings.Builder
	for _, p := range c.Parts {
		if p == nil {
			continue
		}
		if p.Text != "" {
			b.WriteString("--- text ---\n")
			b.WriteString(p.Text)
			b.WriteByte('\n')
		}
		if p.FunctionCall != nil {
			args, _ := json.Marshal(p.FunctionCall.Args)
			b.WriteString("--- function_call ")
			b.WriteString(p.FunctionCall.Name)
			b.WriteString(" ---\n")
			b.Write(args)
			b.WriteByte('\n')
		}
		if p.Thought {
			b.WriteString("--- (thought part omitted) ---\n")
		}
	}
	return strings.TrimSpace(b.String())
}

func runDeepSearchReport(ctx context.Context, db *gorm.DB, deepSearchID uint, userQuery string) (string, error) {
	client, model, err := newGenAIClient(ctx)
	if err != nil {
		return "", err
	}

	sys := `You are a research assistant. The user asked a question; relevant web pages were scraped and summarized into an Intel database.

Your final answer must be a long, detailed markdown report. Aim for depth and length: multiple sections with ## and ### headings, rich paragraphs, and synthesis across sources—not a short summary or bullet-only outline unless the evidence truly supports nothing more. Use the full answer budget when the material warrants it.

Ground every substantive claim in what you retrieved via tools (intel_vector_search, intel_get_content). Call intel_vector_search with focused sub-queries. Whenever search returns Intel ids, call intel_get_content for the most promising ids (especially top hits) before concluding evidence is missing or irrelevant—short vector summaries alone are not enough to decide the database has nothing on a topic.

Structure suggestions (adapt to the question): title line (# heading), executive overview, thematic sections with analysis, a "Sources / evidence" subsection, gaps or uncertainties, and optional brief conclusion. Cite Intel using markdown links only: [short label](intel://<intel_id>) where <intel_id> is the numeric id from tools (same id in every citation to that row). Do not use plain-text forms like "Intel #42". Pull direct quotes or close paraphrases from Intel content where they strengthen the report.

If, after fetching full content for plausible ids, evidence is still thin, say so explicitly—but still write a careful markdown report explaining what was and was not found.

Output markdown only (no JSON wrapper for the final answer).`

	user := strings.TrimSpace(userQuery)
	prompt := "User question:\n\n" + user + "\n\nAfter gathering Intel via tools, produce the final markdown report: comprehensive, well-structured, and as long as the retrieved evidence supports."

	cfg := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(sys, genai.RoleUser),
		Tools:             reportTools(),
		Temperature:       genai.Ptr[float32](0.25),
		MaxOutputTokens:   deepSearchReportMaxOutputTokens(),
	}

	contents := []*genai.Content{genai.NewContentFromText(prompt, genai.RoleUser)}

	for round := 0; round < deepSearchReportMaxToolRounds; round++ {
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("p_seer_deepsearch: report stopped: %w", err)
		}
		op := fmt.Sprintf("deepsearch.report_generate round=%d", round)
		resp, err := p_seer_intel.WithGenAIRetry(ctx, op, func(ctx context.Context) (*genai.GenerateContentResponse, error) {
			return client.Models.GenerateContent(ctx, model, contents, cfg)
		})
		if err != nil {
			slog.Error("p_seer_deepsearch: report generate", "error", err, "round", round)
			return "", fmt.Errorf("p_seer_deepsearch: report generate: %w", err)
		}
		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			return "", fmt.Errorf("p_seer_deepsearch: empty model response (round %d)", round)
		}
		modelContent := resp.Candidates[0].Content
		contents = append(contents, cloneContent(modelContent))

		modelSummary := summarizeModelContentForLog(modelContent)
		respText := strings.TrimSpace(resp.Text())
		fcs := resp.FunctionCalls()

		if deepSearchID != 0 {
			var llmMsg strings.Builder
			llmMsg.WriteString(fmt.Sprintf("round=%d\n", round))
			if modelSummary != "" {
				llmMsg.WriteString("model_content:\n")
				llmMsg.WriteString(truncateRunes(modelSummary, deepSearchReportLogToolPayloadMaxRunes))
				llmMsg.WriteString("\n")
			}
			if respText != "" && respText != modelSummary {
				llmMsg.WriteString("resp.Text() (if distinct):\n")
				llmMsg.WriteString(truncateRunes(respText, deepSearchReportLogToolPayloadMaxRunes))
				llmMsg.WriteString("\n")
			}
			appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindReportLlm, llmMsg.String())
		}

		if len(fcs) == 0 {
			out := strings.TrimSpace(resp.Text())
			if out == "" {
				return "", fmt.Errorf("p_seer_deepsearch: report returned empty text")
			}
			return out, nil
		}

		var parts []*genai.Part
		for _, fc := range fcs {
			if fc == nil {
				continue
			}
			argsJSON, _ := json.Marshal(fc.Args)

			if deepSearchID != 0 {
				appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindReportLlm,
					fmt.Sprintf("round=%d tool=%s args=%s", round, fc.Name, truncateRunes(string(argsJSON), deepSearchReportLogToolPayloadMaxRunes)))
			}

			payload, terr := dispatchReportTool(ctx, db, fc.Name, fc.Args)
			if terr != nil {
				slog.Warn("p_seer_deepsearch: tool error", "tool", fc.Name, "error", terr)
				if deepSearchID != 0 {
					appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindReportLlm,
						fmt.Sprintf("round=%d tool=%s ERROR: %s", round, fc.Name, terr.Error()))
				}
				parts = append(parts, genai.NewPartFromFunctionResponse(fc.Name, map[string]any{"error": terr.Error()}))
				continue
			}
			outJSON, _ := json.Marshal(payload)
			if deepSearchID != 0 {
				appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindReportLlm,
					fmt.Sprintf("round=%d tool=%s result=%s", round, fc.Name, truncateRunes(string(outJSON), deepSearchReportLogToolPayloadMaxRunes)))
			}
			parts = append(parts, genai.NewPartFromFunctionResponse(fc.Name, map[string]any{"output": payload}))
		}
		if len(parts) == 0 {
			return "", fmt.Errorf("p_seer_deepsearch: no tool parts (round %d)", round)
		}
		contents = append(contents, genai.NewContentFromParts(parts, genai.RoleUser))
	}
	return "", fmt.Errorf("p_seer_deepsearch: report exceeded tool rounds")
}

func cloneContent(c *genai.Content) *genai.Content {
	if c == nil {
		return nil
	}
	out := &genai.Content{Role: c.Role, Parts: make([]*genai.Part, len(c.Parts))}
	copy(out.Parts, c.Parts)
	return out
}

func dispatchReportTool(ctx context.Context, db *gorm.DB, name string, args map[string]any) (any, error) {
	switch name {
	case "intel_vector_search":
		q, ok := argString(args, "query")
		if !ok || q == "" {
			return nil, fmt.Errorf("missing query")
		}
		limit := argIntDefault(args, "limit", 8)
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
		id, ok := argUint(args, "intel_id")
		if !ok || id == 0 {
			return nil, fmt.Errorf("missing intel_id")
		}
		doc, err := p_seer_intel.IntelDocumentForTool(ctx, db, id)
		if err != nil {
			return nil, err
		}
		return doc, nil
	default:
		return nil, fmt.Errorf("unknown tool %q", name)
	}
}
