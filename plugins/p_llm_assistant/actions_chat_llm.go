package p_llm_assistant

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_google_genai"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

const assistantSystemPrompt = `You are LLM Assistant inside the Lariv app. You help operators search the public web via Google Programmable Search.

You are a multimodal assistant. You can see, analyze, and process any files, documents, or images attached by the user.

CRITICAL: You have access to various registered skills that help you handle tasks. You MUST check the list of available skills (by calling the list_skills tool) before generating your response to see if an existing skill is suited to the user's request. Checking for available skills is your absolute highest priority.

To properly use a skill, you first need its name, you can get the name using the list_skills tool, then use get_skill_detail to get the content. Content will describe what you need to do with. It will often list rules or a sequence of steps to follow. It may often refer to files, which you can read with read_file. The references files will be listed in the Files section of the response from get_skill_detail.

Even if the task may seem trivial, if a skill might seem to provide some additional information about the task, then you should check the instructions via get_skill_detail.


NOTE: list_skills doesn't give the instructions that are contained in the skill. You NEED to call get_skill_detail to get the instructions.

For normal answers (questions, explanations, summaries after tool results), reply in plain text or markdown.

If a tool response includes an error, explain it briefly and suggest a fix.`

func assistantChatGenConfig(maxOut int) *genai.GenerateContentConfig {
	maxTok := max(int32(maxOut), 1)
	temp := float32(0.35)
	return &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(assistantSystemPrompt, genai.RoleUser),
		Temperature:       &temp,
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
		return nil, nil, fmt.Errorf("p_llm_assistant: empty session")
	}
	last := contents[len(contents)-1]
	if !strings.EqualFold(strings.TrimSpace(last.Role), string(genai.RoleUser)) {
		return nil, nil, fmt.Errorf("p_llm_assistant: last message must be user (got %q)", last.Role)
	}
	if len(last.Parts) == 0 {
		return nil, nil, fmt.Errorf("p_llm_assistant: last user message has no parts")
	}
	history = contents[:len(contents)-1]
	triggerParts = append([]*genai.Part(nil), last.Parts...)
	return history, triggerParts, nil
}

// runAssistantChatStream streams one assistant turn via [genai.Chat.SendStream] (curated history + parts).
func runAssistantChatStream(ctx context.Context, chat *genai.Chat, parts []*genai.Part) (<-chan *genai.Content, <-chan error) {
	contentChan := make(chan *genai.Content)
	errChan := make(chan error, 1)
	go func() {
		defer close(contentChan)
		defer close(errChan)
		if len(parts) == 0 {
			errChan <- fmt.Errorf("p_llm_assistant: empty chat send parts")
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
		errChan <- fmt.Errorf("p_llm_assistant: genai stream retries exhausted")
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
		session, err := gorm.G[LlmAssistantSession](db).Where("id = ?", sessionID).First(ctx)
		if err != nil {
			errChan <- err
			return
		}

		client, err := p_google_genai.NewClient(ctx)
		if err != nil {
			errChan <- err
			return
		}
		model := LlmAssistantPlugin.ChatModel
		maxOut := AssistantAppConfig.ChatMaxOutputTokens
		if maxOut <= 0 {
			maxOut = 1024
		}

		for round := 0; round < maxRounds; round++ {
			if err = ctx.Err(); err != nil {
				errChan <- err
				return
			}

			// Rebuild Chat from DB every round so genai.Chat.curatedHistory never carries a
			// model turn dropped by the SDK's incomplete stream validateContent (which would
			// orphan the next function-response SendStream and trigger API 400).
			contents, err := LoadSessionContents(ctx, db, sessionID)
			if err != nil {
				errChan <- err
				return
			}
			stripDisplayNameFromContents(contents)
			history, triggerParts, err := assistantSplitLastUserContent(contents)
			if err != nil {
				errChan <- err
				return
			}
			chat, err := client.Chats.Create(ctx, model, assistantChatGenConfig(maxOut), history)
			if err != nil {
				errChan <- fmt.Errorf("p_llm_assistant: chats create: %w", err)
				return
			}

			streamChan, streamErrChan := runAssistantChatStream(ctx, chat, triggerParts)
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
					resMap, terr := runToolRound(ctx, db, fc.Name, fc.Args)
					if terr != nil {
						errChan <- terr
						return
					}
					respParts = append(respParts, &genai.Part{
						FunctionResponse: &genai.FunctionResponse{
							ID:       fc.ID,
							Name:     fc.Name,
							Response: resMap,
						},
					})
				}
				if len(respParts) == 0 {
					errChan <- fmt.Errorf("p_llm_assistant: model message had no usable function calls")
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
		var parts []*genai.Part
		for _, p := range src.Parts {
			if !genaiPartIsEmpty(p) {
				parts = append(parts, p)
			}
		}
		return &genai.Content{Role: src.Role, Parts: parts}
	}
	if strings.TrimSpace(dst.Role) == "" {
		dst.Role = src.Role
	}
	for _, p := range src.Parts {
		if !genaiPartIsEmpty(p) {
			dst.Parts = append(dst.Parts, p)
		}
	}
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

func runToolRound(ctx context.Context, db *gorm.DB, name string, args map[string]any) (res map[string]any, err error) {
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tool, ok := LlmToolRegistry.Get(name)
		if !ok {
			res = map[string]any{"error": "unknown tool"}
			return nil
		}
		var runErr error
		res, runErr = tool.Run(ctx, tx, args)
		if runErr != nil {
			res = map[string]any{"error": runErr.Error()}
			return nil
		}
		return nil
	})
	return res, err
}

func stripDisplayNameFromContents(contents []*genai.Content) {
	for _, c := range contents {
		if c == nil {
			continue
		}
		stripDisplayNameFromParts(c.Parts)
	}
}

func stripDisplayNameFromParts(parts []*genai.Part) {
	for _, p := range parts {
		if p == nil {
			continue
		}
		if p.InlineData != nil {
			p.InlineData.DisplayName = ""
		}
	}
}
