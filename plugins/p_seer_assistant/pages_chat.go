package p_seer_assistant

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func registerAssistantMenuPages() {
	lago.RegistryPage.Register("seer_assistant.AssistantMenu", &components.SidebarMenu{
		Title: getters.Static("Assistant"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Chat"),
				Url:   lago.RoutePath("seer_assistant.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("History"),
				Url:   lago.RoutePath("seer_assistant.HistoryRoute", nil),
			},
		},
	})
}

type assistantChatRoot struct {
	components.Page
}

func (e *assistantChatRoot) Build(ctx context.Context) Node {
	wsPath := AppUrl + "ws/"
	sid := assistantOpenSessionID(ctx)
	hiddenVal := "0"
	if sid != 0 {
		hiddenVal = fmt.Sprintf("%d", sid)
	}
	transcriptInner := []Node{}
	if sid != 0 {
		nodes, err := assistantTranscriptNodes(ctx, sid)
		if err != nil {
			transcriptInner = append(transcriptInner, Div(Class("text-error text-sm"), Text("Could not load chat history")))
		} else if len(nodes) > 0 {
			transcriptInner = append(transcriptInner, Group(nodes))
		}
	}
	return Div(
		Class("max-w-3xl mx-auto p-4 flex flex-col gap-4 min-h-[60vh]"),
		Attr("hx-ext", "ws"),
		Attr("ws-connect", wsPath),
		Div(ID("seer_assistant_errors")),
		Div(
			ID("seer_assistant_transcript"),
			Class("flex flex-col gap-2 flex-1 overflow-y-auto border border-base-300 rounded-lg p-3 bg-base-200/40 min-h-[200px]"),
			Group(transcriptInner),
		),
		Div(
			ID("seer_assistant_stream"),
			Class("text-sm font-mono whitespace-pre-wrap min-h-[1.5rem] border border-dashed border-base-300 rounded p-2"),
		),
		FormEl(
			Class("flex flex-col gap-2"),
			Attr("ws-send", ""),
			Input(
				ID("seer_assistant_session_id"),
				Type("hidden"),
				Name("session_id"),
				Value(hiddenVal),
			),
			Textarea(
				Name("message"),
				Class("textarea textarea-bordered w-full"),
				Rows("3"),
				Placeholder("Message…"),
				Required(),
			),
			Button(
				Type("submit"),
				Class("btn btn-primary self-end"),
				Text("Send"),
			),
		),
	)
}

func (e *assistantChatRoot) GetKey() string { return e.Key }

func (e *assistantChatRoot) GetRoles() []string { return e.Roles }

func registerAssistantChatPage() {
	lago.RegistryPage.Register("seer_assistant.ChatPage", &components.ShellScaffold{
		Page: components.Page{Key: "seer_assistant.ChatPage"},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_assistant.AssistantMenu"},
		},
		Children: []components.PageInterface{
			&assistantChatRoot{
				Page: components.Page{Key: "seer_assistant.ChatInner"},
			},
		},
	})
}

func assistantOpenSessionID(ctx context.Context) uint {
	if v := ctx.Value("assistantSession"); v != nil {
		if s, ok := v.(SeerAssistantSession); ok {
			return s.ID
		}
	}
	return 0
}

func assistantTranscriptNodes(ctx context.Context, sessionID uint) ([]Node, error) {
	if sessionID == 0 {
		return nil, nil
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}
	turns, err := BuildChatTurns(ctx, db, sessionID)
	if err != nil {
		return nil, err
	}
	out := make([]Node, 0, len(turns))
	for _, t := range turns {
		// Pass raw text to Text(); gomponents escapes once. Do not pre-escape or
		// apostrophes become visible as &#39; (double-escape).
		body := t.Content
		switch t.Role {
		case "assistant":
			out = append(out, assistantBubbleAssistant(body))
		case "user":
			if strings.HasPrefix(t.Content, "[tool ") {
				out = append(out, assistantBubbleTool(body))
			} else {
				out = append(out, assistantBubbleUser(body))
			}
		default:
			out = append(out, assistantBubbleUser(body))
		}
	}
	return out, nil
}

func assistantBubbleUser(body string) Node {
	return Div(
		Class("chat chat-end mb-2"),
		Div(Class("chat-header text-xs opacity-70"), Text("You")),
		Div(Class("chat-bubble chat-bubble-primary whitespace-pre-wrap"), Text(body)),
	)
}

func assistantBubbleAssistant(body string) Node {
	return Div(
		Class("chat chat-start mb-2"),
		Div(Class("chat-header text-xs opacity-70"), Text("Assistant")),
		Div(Class("chat-bubble chat-bubble-secondary whitespace-pre-wrap"), Text(body)),
	)
}

func assistantBubbleTool(body string) Node {
	return Div(
		Class("chat chat-start mb-2"),
		Div(Class("chat-header text-xs opacity-70"), Text("Tool")),
		Div(Class("chat-bubble chat-bubble-accent text-sm whitespace-pre-wrap"), Text(body)),
	)
}

func init() {
	registerAssistantMenuPages()
	registerAssistantChatPage()
}
