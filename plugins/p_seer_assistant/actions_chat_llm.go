package p_seer_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"

	"github.com/lariv-in/lago/plugins/p_google_genai"
	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"github.com/lariv-in/lago/plugins/p_seer_reddit"
	"golang.org/x/net/websocket"
	"gorm.io/gorm"
)

const assistantSystemPrompt = `You are Seer Assistant inside the Lago app. You help operators manage Reddit ingestion, Seer Websites (crawl sources + workers), search Intel (vector summaries), and the public web (Google).

When you need data or to perform an action, reply with a single JSON object and nothing else (no markdown fences, no commentary). Use one of:
- {"tool":"intel_search","query":"<natural language query>","limit":<1-20>}
- {"tool":"google_search","query":"<web search query>","limit":<1-20>}
- {"tool":"reddit_add_source","subreddits":["nameWithoutRPrefix"],"search_query":"","max_fresh_posts":<uint optional>,"load_websites":<bool>,"reddit_runner_id":<optional; omit key entirely to leave worker unset; if present must match an existing runner id>}
- {"tool":"reddit_edit_source","reddit_source_id":<uint>,"subreddits":["..."],"search_query":"","max_fresh_posts":<uint>,"load_websites":<bool>,"reddit_runner_id":<optional same as add>}
- {"tool":"reddit_edit_worker","reddit_runner_id":<uint>,"worker_name":"<label>","worker_duration":"<Go duration e.g. 1h, 30m, 90s>"}
- {"tool":"website_list_sources"}
- {"tool":"website_list_workers"}
- {"tool":"website_add_source","seed_url":"<https URL>","website_depth":<uint; link hops after seed, 0=seed only>,"website_runner_id":<optional uint; omit to leave worker unset>}
- {"tool":"website_edit_source","website_source_id":<uint>,"seed_url":"<https URL>","website_depth":<uint>,"website_runner_id":<optional>}
- {"tool":"website_add_worker","worker_name":"<label>","worker_duration":"<Go duration e.g. 1h, 30m>"}
- {"tool":"website_edit_worker","website_runner_id":<uint>,"worker_name":"<label>","worker_duration":"<Go duration>"}

Before reddit_add_source: run google_search first when you need to discover or verify subreddit names, topics, or anything else best confirmed on the web (then use findings in the next reddit_add_source call).

For normal answers (questions, explanations, summaries of prior tool results), reply in plain text or markdown — not JSON.

If a tool returns an error, explain it briefly and suggest a fix.`

type assistantToolEnvelope struct {
	Tool           string   `json:"tool"`
	Query          string   `json:"query,omitempty"`
	Limit          int      `json:"limit,omitempty"`
	RedditRunnerID *uint    `json:"reddit_runner_id,omitempty"`
	Subreddits     []string `json:"subreddits,omitempty"`
	SearchQuery    string   `json:"search_query,omitempty"`
	MaxFreshPosts  uint     `json:"max_fresh_posts,omitempty"`
	LoadWebsites   bool     `json:"load_websites,omitempty"`
	RedditSourceID uint     `json:"reddit_source_id,omitempty"`
	WorkerName     string   `json:"worker_name,omitempty"`
	WorkerDuration string   `json:"worker_duration,omitempty"`
	// Seer Websites (p_seer_websites); website_runner_id optional on add/edit source, required on website_edit_worker.
	SeedURL         string `json:"seed_url,omitempty"`
	WebsiteDepth    uint   `json:"website_depth,omitempty"`
	WebsiteSourceID uint   `json:"website_source_id,omitempty"`
	WebsiteRunnerFK *uint  `json:"website_runner_id,omitempty"`
}

func floatPtr(f float32) *float32 { return &f }

func runAssistantLLMStream(ctx context.Context, systemPrompt string, turns []AssistantChatTurn, maxOut int, onToken func(string) error) (string, error) {
	gen := p_google_genai.GenerateRequest{
		SystemPrompt:    systemPrompt,
		UserPrompt:      "",
		Temperature:     floatPtr(0.35),
		MaxOutputTokens: maxOut,
		Thinking:        &p_google_genai.ThinkingConfig{Mode: p_google_genai.ThinkingModeDisabled},
	}
	gg := make([]p_google_genai.ChatTurn, len(turns))
	for i := range turns {
		gg[i] = p_google_genai.ChatTurn{Role: turns[i].Role, Content: turns[i].Content}
	}
	return p_google_genai.GenerateChatStream(ctx, gen, gg, onToken)
}

// RunAssistantAfterUserMessage runs tool rounds + streaming assistant output on ws.
func RunAssistantAfterUserMessage(ctx context.Context, db *gorm.DB, ws *websocket.Conn, sessionID uint) error {
	maxRounds := AssistantAppConfig.AssistantToolRounds
	if maxRounds <= 0 {
		maxRounds = 8
	}
	for round := 0; round < maxRounds; round++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		turns, err := BuildChatTurns(ctx, db, sessionID)
		if err != nil {
			return err
		}
		maxOut := AssistantAppConfig.ChatMaxOutputTokens
		if maxOut <= 0 {
			maxOut = 1024
		}

		tokenCh := make(chan string, 128)
		streamErr := make(chan error, 1)
		go func() {
			var streamWriteErr error
			defer func() { streamErr <- streamWriteErr }()
			for piece := range tokenCh {
				frag := fmt.Sprintf(`<div id="seer_assistant_stream" hx-swap-oob="beforeend">%s</div>`, html.EscapeString(piece))
				if _, err := ws.Write([]byte(frag)); err != nil {
					streamWriteErr = err
					return
				}
			}
		}()

		full, err := runAssistantLLMStream(ctx, assistantSystemPrompt, turns, maxOut, func(s string) error {
			select {
			case tokenCh <- s:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		close(tokenCh)
		if werr := <-streamErr; werr != nil {
			return werr
		}
		if err != nil {
			_ = writeWSHTML(ws, errorOOB(err))
			return err
		}

		if env, ok := parseToolEnvelope(full); ok {
			if err := writeWSHTML(ws, `<div id="seer_assistant_stream" hx-swap-oob="true"></div>`); err != nil {
				return err
			}
			if err := runPersistedToolRound(ctx, db, ws, sessionID, env); err != nil {
				return err
			}
			continue
		}

		if err := writeWSHTML(ws, `<div id="seer_assistant_stream" hx-swap-oob="true"></div>`); err != nil {
			return err
		}
		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			_, err := AppendAssistantMessage(ctx, tx, sessionID, full)
			return err
		}); err != nil {
			_ = writeWSHTML(ws, errorOOB(err))
			return err
		}
		escaped := html.EscapeString(full)
		bubble := fmt.Sprintf(
			`<div id="seer_assistant_transcript" hx-swap-oob="beforeend"><div class="chat chat-start mb-2"><div class="chat-header text-xs opacity-70">Assistant</div><div class="chat-bubble chat-bubble-secondary whitespace-pre-wrap">%s</div></div></div>`,
			escaped,
		)
		return writeWSHTML(ws, bubble)
	}
	return writeWSHTML(ws, errorOOB(fmt.Errorf("assistant: tool round limit exceeded")))
}

func writeWSHTML(ws *websocket.Conn, s string) error {
	_, err := ws.Write([]byte(s))
	return err
}

func errorOOB(err error) string {
	msg := html.EscapeString(err.Error())
	return fmt.Sprintf(
		`<div id="seer_assistant_errors" hx-swap-oob="true"><div class="alert alert-error text-sm">%s</div></div>`,
		msg,
	)
}

func parseToolEnvelope(s string) (*assistantToolEnvelope, bool) {
	raw := strings.TrimSpace(s)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "{") {
		return nil, false
	}
	var env assistantToolEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		return nil, false
	}
	switch strings.TrimSpace(strings.ToLower(env.Tool)) {
	case "intel_search", "google_search",
		"reddit_add_source", "reddit_edit_source", "reddit_edit_worker",
		"website_list_sources", "website_list_workers",
		"website_add_source", "website_edit_source",
		"website_add_worker", "website_edit_worker":
		return &env, true
	default:
		return nil, false
	}
}

func runPersistedToolRound(ctx context.Context, db *gorm.DB, ws *websocket.Conn, sessionID uint, env *assistantToolEnvelope) error {
	var call SeerAssistantToolCall
	var resText string
	var errText string

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ierr error
		call, ierr = AppendToolCall(ctx, tx, sessionID, env.Tool, env)
		if ierr != nil {
			return ierr
		}
		switch strings.TrimSpace(strings.ToLower(env.Tool)) {
		case "intel_search":
			resText, errText = runIntelSearchTool(ctx, db, env)
		case "google_search":
			resText, errText = runGoogleSearchTool(ctx, env)
		case "reddit_add_source":
			resText, errText = runRedditAddSourceTool(ctx, tx, env)
		case "reddit_edit_source":
			resText, errText = runRedditEditSourceTool(ctx, tx, env)
		case "reddit_edit_worker":
			resText, errText = runRedditEditWorkerTool(ctx, tx, env)
		case "website_list_sources":
			resText, errText = runWebsiteListSourcesTool(ctx, db, env)
		case "website_list_workers":
			resText, errText = runWebsiteListWorkersTool(ctx, db, env)
		case "website_add_source":
			resText, errText = runWebsiteAddSourceTool(ctx, tx, env)
		case "website_edit_source":
			resText, errText = runWebsiteEditSourceTool(ctx, tx, env)
		case "website_add_worker":
			resText, errText = runWebsiteAddWorkerTool(ctx, tx, env)
		case "website_edit_worker":
			resText, errText = runWebsiteEditWorkerTool(ctx, tx, env)
		default:
			errText = "unknown tool"
		}
		_, ierr = AppendToolResult(ctx, tx, sessionID, call.ID, resText, errText)
		return ierr
	})
	if err != nil {
		slog.Error("p_seer_assistant: tool round", "error", err)
		_ = writeWSHTML(ws, errorOOB(err))
		return err
	}
	note := fmt.Sprintf("Tool %q finished.", env.Tool)
	if errText != "" {
		note = fmt.Sprintf("Tool %q error: %s", env.Tool, errText)
	}
	frag := fmt.Sprintf(
		`<div id="seer_assistant_transcript" hx-swap-oob="beforeend"><div class="chat chat-start mb-2"><div class="chat-header text-xs opacity-70">Tool</div><div class="chat-bubble chat-bubble-accent text-sm whitespace-pre-wrap">%s</div></div></div>`,
		html.EscapeString(note),
	)
	return writeWSHTML(ws, frag)
}

func runIntelSearchTool(ctx context.Context, db *gorm.DB, env *assistantToolEnvelope) (string, string) {
	q := strings.TrimSpace(env.Query)
	if q == "" {
		return "", "intel_search: empty query"
	}
	limit := env.Limit
	if limit <= 0 {
		limit = 8
	}
	if limit > AssistantAppConfig.IntelSearchLimitCap {
		limit = AssistantAppConfig.IntelSearchLimitCap
	}
	rows, err := p_seer_intel.SearchIntelBySimilarity(ctx, db, q, limit)
	if err != nil {
		return "", err.Error()
	}
	if len(rows) == 0 {
		return "[]", ""
	}
	type row struct {
		ID      uint   `json:"id"`
		Title   string `json:"title"`
		Summary string `json:"summary"`
		Kind    string `json:"kind"`
		KindID  uint   `json:"kind_id"`
	}
	out := make([]row, 0, len(rows))
	for _, r := range rows {
		out = append(out, row{ID: r.ID, Title: r.Title, Summary: r.Summary, Kind: r.Kind, KindID: r.KindID})
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}

func runRedditAddSourceTool(ctx context.Context, tx *gorm.DB, env *assistantToolEnvelope) (string, string) {
	var runner *uint
	if env.RedditRunnerID != nil && *env.RedditRunnerID != 0 {
		id := *env.RedditRunnerID
		runner = &id
	}
	p := p_seer_reddit.RedditSourceCreateParams{
		RedditRunnerID: runner,
		Subreddits:     env.Subreddits,
		SearchQuery:    env.SearchQuery,
		MaxFreshPosts:  env.MaxFreshPosts,
		LoadWebsites:   env.LoadWebsites,
	}
	src, err := p_seer_reddit.CreateRedditSource(ctx, tx, p)
	if err != nil {
		return "", err.Error()
	}
	return fmt.Sprintf(`{"reddit_source_id":%d}`, src.ID), ""
}

func runRedditEditSourceTool(ctx context.Context, tx *gorm.DB, env *assistantToolEnvelope) (string, string) {
	if env.RedditSourceID == 0 {
		return "", "reddit_edit_source: reddit_source_id required"
	}
	var runner *uint
	if env.RedditRunnerID != nil && *env.RedditRunnerID != 0 {
		id := *env.RedditRunnerID
		runner = &id
	}
	p := p_seer_reddit.RedditSourceUpdateParams{
		SourceID: env.RedditSourceID,
		RedditSourceCreateParams: p_seer_reddit.RedditSourceCreateParams{
			RedditRunnerID: runner,
			Subreddits:     env.Subreddits,
			SearchQuery:    env.SearchQuery,
			MaxFreshPosts:  env.MaxFreshPosts,
			LoadWebsites:   env.LoadWebsites,
		},
	}
	src, err := p_seer_reddit.UpdateRedditSource(ctx, tx, p)
	if err != nil {
		return "", err.Error()
	}
	return fmt.Sprintf(`{"reddit_source_id":%d}`, src.ID), ""
}

func runRedditEditWorkerTool(ctx context.Context, tx *gorm.DB, env *assistantToolEnvelope) (string, string) {
	var rid uint
	if env.RedditRunnerID != nil && *env.RedditRunnerID != 0 {
		rid = *env.RedditRunnerID
	}
	if rid == 0 {
		return "", "reddit_edit_worker: reddit_runner_id required"
	}
	name := strings.TrimSpace(env.WorkerName)
	if name == "" {
		return "", "reddit_edit_worker: worker_name required"
	}
	durStr := strings.TrimSpace(env.WorkerDuration)
	if durStr == "" {
		return "", "reddit_edit_worker: worker_duration required (Go duration, e.g. 1h, 45m, 90s)"
	}
	d, err := time.ParseDuration(durStr)
	if err != nil {
		return "", fmt.Sprintf("reddit_edit_worker: invalid worker_duration: %v", err)
	}
	runner, err := p_seer_reddit.UpdateRedditRunner(ctx, tx, p_seer_reddit.RedditRunnerUpdateParams{
		ID:       rid,
		Name:     name,
		Duration: d,
	})
	if err != nil {
		return "", err.Error()
	}
	out := map[string]any{
		"reddit_runner_id": runner.ID,
		"name":             runner.Name,
		"duration":         runner.Duration.String(),
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}
