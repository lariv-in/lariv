package p_seer_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_google_genai"
	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"github.com/lariv-in/lago/plugins/p_seer_reddit"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

const assistantSystemPrompt = `You are Seer Assistant inside the Lago app. You help operators manage Reddit ingestion, Seer Websites (crawl sources + workers), search Intel (vector summaries), and the public web (Google).

Use the declared tools when the user needs data or actions from those systems. Before reddit_add_source, call google_search when you need to discover or verify subreddit names or topics on the web.

For normal answers (questions, explanations, summaries after tool results), reply in plain text or markdown.

If a tool response includes an error, explain it briefly and suggest a fix.`

func assistantChatGenConfig(maxOut int) *genai.GenerateContentConfig {
	maxTok := max(int32(maxOut), 1)
	return &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(assistantSystemPrompt, genai.RoleUser),
		Temperature:       new(float32(0.35)),
		MaxOutputTokens:   maxTok,
		Tools:             assistantGeminiTools(),
		ToolConfig: &genai.ToolConfig{
			FunctionCallingConfig: &genai.FunctionCallingConfig{
				Mode: genai.FunctionCallingConfigModeAuto,
			},
		},
	}
}

// assistantSplitLastUserContent returns history for [genai.Chats.Create] and the trailing user [genai.Part]
// slice for the first [genai.Chat.SendStream] (the latest user turn must not be duplicated in history).
func assistantSplitLastUserContent(contents []*genai.Content) (history []*genai.Content, triggerParts []*genai.Part, err error) {
	if len(contents) == 0 {
		return nil, nil, fmt.Errorf("p_seer_assistant: empty session")
	}
	last := contents[len(contents)-1]
	if !strings.EqualFold(strings.TrimSpace(last.Role), string(genai.RoleUser)) {
		return nil, nil, fmt.Errorf("p_seer_assistant: last message must be user (got %q)", last.Role)
	}
	if len(last.Parts) == 0 {
		return nil, nil, fmt.Errorf("p_seer_assistant: last user message has no parts")
	}
	history = contents[:len(contents)-1]
	triggerParts = append([]*genai.Part(nil), last.Parts...)
	return history, triggerParts, nil
}

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

// runAssistantChatStream streams one assistant turn via [genai.Chat.SendStream] (curated history + parts).
func runAssistantChatStream(ctx context.Context, chat *genai.Chat, parts []*genai.Part) (<-chan *genai.Content, <-chan error) {
	contentChan := make(chan *genai.Content)
	errChan := make(chan error, 1)
	go func() {
		defer close(contentChan)
		defer close(errChan)
		if len(parts) == 0 {
			errChan <- fmt.Errorf("p_seer_assistant: empty chat send parts")
			return
		}
		for attempt := 0; attempt < p_google_genai.DefaultStreamMaxAttempts; attempt++ {
			if attempt > 0 {
				backoff := time.Duration(500*(1<<uint(attempt-1))) * time.Millisecond
				if backoff > 12*time.Second {
					backoff = 12 * time.Second
				}
				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				}
			}
			emittedChunks := 0
			retryLater := false
			for resp, err := range chat.SendStream(ctx, parts...) {
				if err != nil {
					if emittedChunks == 0 && p_google_genai.RetryableQuotaError(err) && attempt < p_google_genai.DefaultStreamMaxAttempts-1 {
						retryLater = true
						break
					}
					errChan <- err
					return
				}
				if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
					continue
				}
				piece := cloneGenAIContent(resp.Candidates[0].Content)
				if piece == nil {
					continue
				}
				emittedChunks++
				select {
				case contentChan <- piece:
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				}
			}
			if retryLater {
				continue
			}
			return
		}
		errChan <- fmt.Errorf("p_seer_assistant: genai stream retries exhausted")
	}()
	return contentChan, errChan
}

func cloneGenAIContent(c *genai.Content) *genai.Content {
	if c == nil {
		return nil
	}
	parts := make([]*genai.Part, len(c.Parts))
	copy(parts, c.Parts)
	return &genai.Content{Role: c.Role, Parts: parts}
}

func RunAssistant(r *http.Request, sessionID uint) (chan *genai.Content, chan error) {
	contentChan, errChan := make(chan *genai.Content), make(chan error, 1)
	maxRounds := AssistantAppConfig.AssistantToolRounds
	if maxRounds <= 0 {
		maxRounds = 8
	}
	ctx := r.Context()
	go func() {
		defer close(contentChan)
		defer close(errChan)

		db, err := getters.DBFromContext(ctx)
		if err != nil {
			errChan <- err
			return
		}
		session, err := gorm.G[SeerAssistantSession](db).Where("id = ?", sessionID).First(ctx)
		if err != nil {
			errChan <- err
			return
		}

		client, err := p_google_genai.NewClient(ctx)
		if err != nil {
			errChan <- err
			return
		}
		model := SeerAssistantPlugin.ChatModel
		maxOut := AssistantAppConfig.ChatMaxOutputTokens
		if maxOut <= 0 {
			maxOut = 1024
		}
		contents, err := LoadSessionContents(ctx, db, sessionID)
		if err != nil {
			errChan <- err
			return
		}
		history, triggerParts, err := assistantSplitLastUserContent(contents)
		if err != nil {
			errChan <- err
			return
		}
		chat, err := client.Chats.Create(ctx, model, assistantChatGenConfig(maxOut), history)
		if err != nil {
			errChan <- fmt.Errorf("p_seer_assistant: chats create: %w", err)
			return
		}
		nextParts := triggerParts

		for round := 0; round < maxRounds; round++ {
			if err = ctx.Err(); err != nil {
				errChan <- err
				return
			}

			streamChan, streamErrChan := runAssistantChatStream(ctx, chat, nextParts)
			var full *genai.Content
			for streamChan != nil || streamErrChan != nil {
				select {
				case piece, ok := <-streamChan:
					if !ok {
						streamChan = nil
						continue
					}
					piece = normalizeAssistantContent(piece)
					full = mergeAssistantContent(full, piece)
					select {
					case contentChan <- piece:
					case <-ctx.Done():
						errChan <- ctx.Err()
						return
					}
				case err, ok := <-streamErrChan:
					if !ok {
						streamErrChan = nil
						continue
					}
					if err != nil {
						errChan <- err
						return
					}
				}
			}

			if full != nil && assistantContentHasFunctionCall(full) {
				if err := session.SaveContent(ctx, *full); err != nil {
					errChan <- err
					return
				}
				var respParts []*genai.Part
				for _, part := range full.Parts {
					if part == nil || part.FunctionCall == nil {
						continue
					}
					fc := part.FunctionCall
					env, aerr := assistantEnvFromFunctionCall(fc)
					if aerr != nil {
						errChan <- aerr
						return
					}
					resText, errText, terr := runToolRound(ctx, db, env)
					if terr != nil {
						errChan <- terr
						return
					}
					respParts = append(respParts, &genai.Part{
						FunctionResponse: &genai.FunctionResponse{
							ID:       fc.ID,
							Name:     fc.Name,
							Response: assistantFunctionResultMap(resText, errText),
						},
					})
				}
				if len(respParts) == 0 {
					errChan <- fmt.Errorf("p_seer_assistant: model message had no usable function calls")
					return
				}
				userTool := &genai.Content{Role: genai.RoleUser, Parts: respParts}
				if err := session.SaveContent(ctx, *userTool); err != nil {
					errChan <- err
					return
				}
				select {
				case contentChan <- userTool:
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				}
				nextParts = respParts
				continue
			}

			if full != nil {
				err = session.SaveContent(ctx, *full)
				if err != nil {
					errChan <- err
					return
				}
			}
			return
		}
		errChan <- fmt.Errorf("assistant: tool round limit exceeded")
	}()
	return contentChan, errChan
}

func normalizeAssistantContent(content *genai.Content) *genai.Content {
	if content == nil {
		return nil
	}
	if strings.TrimSpace(content.Role) == "" {
		content.Role = genai.RoleModel
	}
	return content
}

func mergeAssistantContent(dst, src *genai.Content) *genai.Content {
	if src == nil {
		return dst
	}
	if dst == nil {
		clone := &genai.Content{
			Role:  src.Role,
			Parts: append([]*genai.Part(nil), src.Parts...),
		}
		return clone
	}
	if strings.TrimSpace(dst.Role) == "" {
		dst.Role = src.Role
	}
	dst.Parts = append(dst.Parts, src.Parts...)
	return dst
}

func assistantContentHasFunctionCall(c *genai.Content) bool {
	if c == nil {
		return false
	}
	for _, part := range c.Parts {
		if part != nil && part.FunctionCall != nil {
			return true
		}
	}
	return false
}

func assistantAllowedToolName(name string) bool {
	switch strings.TrimSpace(strings.ToLower(name)) {
	case "intel_search", "google_search",
		"reddit_add_source", "reddit_edit_source",
		"reddit_add_worker", "reddit_edit_worker",
		"website_list_sources", "website_list_workers",
		"website_add_source", "website_edit_source",
		"website_add_worker", "website_edit_worker":
		return true
	default:
		return false
	}
}

func assistantEnvFromFunctionCall(fc *genai.FunctionCall) (*assistantToolEnvelope, error) {
	if fc == nil || fc.Name == "" {
		return nil, fmt.Errorf("p_seer_assistant: empty function call")
	}
	name := strings.TrimSpace(strings.ToLower(fc.Name))
	if !assistantAllowedToolName(name) {
		return nil, fmt.Errorf("p_seer_assistant: unknown tool %q", fc.Name)
	}
	env := &assistantToolEnvelope{Tool: name}
	if fc.Args != nil {
		raw, err := json.Marshal(fc.Args)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, env); err != nil {
			return nil, fmt.Errorf("p_seer_assistant: decode tool args: %w", err)
		}
	}
	env.Tool = name
	return env, nil
}

func assistantFunctionResultMap(resText, errText string) map[string]any {
	if strings.TrimSpace(errText) != "" {
		return map[string]any{"error": errText}
	}
	if strings.TrimSpace(resText) == "" {
		return map[string]any{"output": ""}
	}
	var parsed any
	if err := json.Unmarshal([]byte(resText), &parsed); err == nil {
		return map[string]any{"output": parsed}
	}
	return map[string]any{"output": resText}
}

func runToolRound(ctx context.Context, db *gorm.DB, env *assistantToolEnvelope) (resText string, errText string, err error) {
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		switch strings.TrimSpace(strings.ToLower(env.Tool)) {
		case "intel_search":
			resText, errText = runIntelSearchTool(ctx, tx, env)
		case "google_search":
			resText, errText = runGoogleSearchTool(ctx, env)
		case "reddit_add_source":
			resText, errText = runRedditAddSourceTool(ctx, tx, env)
		case "reddit_edit_source":
			resText, errText = runRedditEditSourceTool(ctx, tx, env)
		case "reddit_add_worker":
			resText, errText = runRedditAddWorkerTool(ctx, tx, env)
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
		return nil
	})
	return resText, errText, err
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

func runRedditAddWorkerTool(ctx context.Context, tx *gorm.DB, env *assistantToolEnvelope) (string, string) {
	name := strings.TrimSpace(env.WorkerName)
	if name == "" {
		return "", "reddit_add_worker: worker_name required"
	}
	durStr := strings.TrimSpace(env.WorkerDuration)
	if durStr == "" {
		return "", "reddit_add_worker: worker_duration required (Go duration, e.g. 1h, 45m, 90s)"
	}
	d, err := time.ParseDuration(durStr)
	if err != nil {
		return "", fmt.Sprintf("reddit_add_worker: invalid worker_duration: %v", err)
	}
	runner, err := p_seer_reddit.CreateRedditRunner(ctx, tx, p_seer_reddit.RedditRunnerCreateParams{
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
