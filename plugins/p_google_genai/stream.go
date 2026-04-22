package p_google_genai

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// ChatTurn is one message in a multi-turn chat completion.
type ChatTurn struct {
	Role    string
	Content string
}

// GenerateChatStream runs a multi-turn chat with optional streaming tokens.
func GenerateChatStream(ctx context.Context, req GenerateRequest, messages []ChatTurn, onToken func(string) error) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("p_google_genai: GenerateChatStream: no messages")
	}
	cli, err := genaiClientFor(ctx)
	if err != nil {
		return "", err
	}
	sys, contents, err := buildChatContents(req, messages)
	if err != nil {
		return "", err
	}
	cfg := baseGenerateConfig(req)
	if sys != "" {
		cfg.SystemInstruction = genai.NewContentFromText(sys, "")
	}
	attachExplicitContextCache(ctx, cli, GoogleGenAIConfig.TextModel, cfg)
	return runGenerateStream(ctx, cli, GoogleGenAIConfig.TextModel, contents, cfg, onToken)
}

func buildChatContents(req GenerateRequest, messages []ChatTurn) (system string, contents []*genai.Content, err error) {
	var sysParts []string
	if sp := strings.TrimSpace(req.SystemPrompt); sp != "" {
		sysParts = append(sysParts, applyThinkingDirective(sp, resolveThinking(req.Thinking)))
	} else {
		sysParts = append(sysParts, applyThinkingDirective("", resolveThinking(req.Thinking)))
	}
	for _, m := range messages {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		text := strings.TrimSpace(m.Content)
		if role == "" {
			return "", nil, fmt.Errorf("p_google_genai: GenerateChatStream: empty role")
		}
		switch role {
		case "system":
			if text != "" {
				sysParts = append(sysParts, text)
			}
		case "assistant", "model":
			contents = append(contents, genai.NewContentFromText(text, genai.RoleModel))
		default:
			contents = append(contents, genai.NewContentFromText(text, genai.RoleUser))
		}
	}
	system = strings.TrimSpace(strings.Join(sysParts, "\n\n"))
	if len(contents) == 0 {
		return "", nil, fmt.Errorf("p_google_genai: GenerateChatStream: no user/model turns")
	}
	return system, contents, nil
}
