package sqlagent

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

const (
	adkAppName        = "p_sqlagent"
	sqlAgentName      = "sqlagent_chat"
	defaultGeminiModel = "gemini-3.1-flash-lite-preview"
)

type adkRuntime struct {
	runner   *runner.Runner
	sessions session.Service
}

var (
	adkMu     sync.Mutex
	adkLoaded bool
	adkRT     *adkRuntime
	adkInitErr error
)

func geminiAPIKey() string {
	if k := strings.TrimSpace(os.Getenv("GOOGLE_API_KEY")); k != "" {
		return k
	}
	return strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
}

func geminiModelID() string {
	if m := strings.TrimSpace(os.Getenv("SQLAGENT_GEMINI_MODEL")); m != "" {
		return m
	}
	return defaultGeminiModel
}

func loadADK(ctx context.Context) (*adkRuntime, error) {
	adkMu.Lock()
	defer adkMu.Unlock()
	if adkLoaded {
		return adkRT, adkInitErr
	}
	adkLoaded = true
	key := geminiAPIKey()
	if key == "" {
		adkInitErr = errors.New("no Gemini API key: set GOOGLE_API_KEY or GEMINI_API_KEY")
		logError("sqlagent: load ADK", adkInitErr)
		return nil, adkInitErr
	}
	modelID := geminiModelID()
	m, err := gemini.NewModel(ctx, modelID, &genai.ClientConfig{APIKey: key})
	if err != nil {
		adkInitErr = err
		logError("sqlagent: gemini NewModel", err, "model", modelID)
		return nil, err
	}
	sqlT, err := newSQLTool()
	if err != nil {
		adkInitErr = err
		return nil, err
	}
	a, err := llmagent.New(llmagent.Config{
		Name:        sqlAgentName,
		Model:       m,
		Description: "Assistant that discusses SQL and database tasks and can run raw SQL via a single tool.",
		Instruction: `You are a helpful assistant embedded in a SQL agent chat UI.
Be concise. You may explain SQL concepts and suggest query ideas.
You have a tool "` + sqlToolName + `" that runs one raw SQL string per call on the same database transaction as the chat turn. Read that tool's description for input shape and result format.`,
		Tools: []tool.Tool{sqlT},
	})
	if err != nil {
		adkInitErr = err
		logError("sqlagent: llmagent New", err)
		return nil, err
	}
	sessSvc := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:           adkAppName,
		Agent:             a,
		SessionService:    sessSvc,
		AutoCreateSession: true,
	})
	if err != nil {
		adkInitErr = err
		logError("sqlagent: runner New", err)
		return nil, err
	}
	adkRT = &adkRuntime{runner: r, sessions: sessSvc}
	return adkRT, nil
}

func genaiTextContent(c *genai.Content) string {
	if c == nil {
		return ""
	}
	var b strings.Builder
	for _, p := range c.Parts {
		if p == nil || p.Thought {
			continue
		}
		if p.Text != "" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

// seedADKSessionFromDB loads prior turns from the database into a cold ADK session
// (in-memory only) so restarts still see conversation context.
func seedADKSessionFromDB(ctx context.Context, rt *adkRuntime, db *gorm.DB, userID, conversationID uint, beforeSortOrder int) error {
	uid := strconv.FormatUint(uint64(userID), 10)
	sid := strconv.FormatUint(uint64(conversationID), 10)

	getResp, err := rt.sessions.Get(ctx, &session.GetRequest{AppName: adkAppName, UserID: uid, SessionID: sid})
	if err != nil {
		if _, cerr := rt.sessions.Create(ctx, &session.CreateRequest{AppName: adkAppName, UserID: uid, SessionID: sid}); cerr != nil {
			logError("sqlagent: ADK session Create", cerr, "session_id", sid)
			return cerr
		}
		getResp, err = rt.sessions.Get(ctx, &session.GetRequest{AppName: adkAppName, UserID: uid, SessionID: sid})
		if err != nil {
			logError("sqlagent: ADK session Get after Create", err, "session_id", sid)
			return err
		}
	}
	sess := getResp.Session
	if sess.Events().Len() > 0 {
		return nil
	}

	msgs, err := LoadMessagesForConversation(db, conversationID)
	if err != nil {
		logError("sqlagent: LoadMessagesForConversation (seed ADK)", err, "conversation_id", conversationID)
		return err
	}
	for i := range msgs {
		m := &msgs[i]
		if m.SortOrder >= beforeSortOrder {
			continue
		}
		switch m.Kind {
		case MessageKindUser:
			if m.UserMessage == nil {
				continue
			}
			content := m.UserMessage.Content
			if isRegistrySchemaBootstrapUserContent(content) {
				body := strings.TrimSpace(strings.TrimPrefix(content, registrySchemaUserMessagePrefix))
				if body == "" {
					continue
				}
				content = "[Registry schema — internal context]\n" + body
			}
			ev := session.NewEvent("db-seed")
			ev.Author = "user"
			ev.LLMResponse = model.LLMResponse{Content: genai.NewContentFromText(content, genai.RoleUser)}
			if err := rt.sessions.AppendEvent(ctx, sess, ev); err != nil {
				logError("sqlagent: ADK AppendEvent (user seed)", err, "conversation_id", conversationID)
				return err
			}
		case MessageKindTool:
			if !isRegistrySchemaToolMessage(m) {
				continue
			}
			body := strings.TrimSpace(m.ToolMessage.Detail)
			if body == "" {
				continue
			}
			// Hidden from UI; inject as user-role context for the model.
			text := "[Registry schema — internal context]\n" + body
			ev := session.NewEvent("db-seed")
			ev.Author = "user"
			ev.LLMResponse = model.LLMResponse{Content: genai.NewContentFromText(text, genai.RoleUser)}
			if err := rt.sessions.AppendEvent(ctx, sess, ev); err != nil {
				logError("sqlagent: ADK AppendEvent (tool seed)", err, "conversation_id", conversationID)
				return err
			}
		case MessageKindAI:
			if m.AIMessage == nil || m.AIMessage.Status != AIStatusComplete {
				continue
			}
			body := strings.TrimSpace(m.AIMessage.Content)
			if body == "" {
				continue
			}
			ev := session.NewEvent("db-seed")
			ev.Author = sqlAgentName
			ev.LLMResponse = model.LLMResponse{Content: genai.NewContentFromText(body, genai.RoleModel)}
			if err := rt.sessions.AppendEvent(ctx, sess, ev); err != nil {
				logError("sqlagent: ADK AppendEvent (ai seed)", err, "conversation_id", conversationID)
				return err
			}
		default:
			continue
		}
	}
	return nil
}

// ForEachADKReplyChunk runs one ADK turn with SSE streaming and invokes fn for each model chunk
// (cumulative text). fn may be called with empty text on some events; callers typically skip those.
func ForEachADKReplyChunk(
	ctx context.Context,
	rt *adkRuntime,
	userID, conversationID uint,
	userText string,
	fn func(text string) error,
) error {
	uid := strconv.FormatUint(uint64(userID), 10)
	sid := strconv.FormatUint(uint64(conversationID), 10)
	msg := genai.NewContentFromText(userText, genai.RoleUser)
	for ev, err := range rt.runner.Run(ctx, uid, sid, msg, agent.RunConfig{StreamingMode: agent.StreamingModeSSE}) {
		if err != nil {
			logError("sqlagent: runner Run event", err, "session_id", sid)
			return err
		}
		if ev == nil || ev.Author == "user" {
			continue
		}
		if ev.Author != sqlAgentName {
			continue
		}
		text := genaiTextContent(ev.LLMResponse.Content)
		if err := fn(text); err != nil {
			logError("sqlagent: ADK reply chunk callback", err, "session_id", sid)
			return err
		}
	}
	return nil
}
