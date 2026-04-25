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
	sid := assistantOpenSessionID(ctx)
	wsPath := AppUrl + "ws/"
	if sid != 0 {
		wsPath = fmt.Sprintf("%s?session_id=%d", wsPath, sid)
	}
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
		Script(Raw(`document.body.addEventListener("htmx:wsConfigSend", function(event) {
  if (!event || !event.detail || !event.detail.parameters) {
    return;
  }
  if (!event.target || event.target.id !== "seer_assistant_chat_form") {
    return;
  }
  var raw = event.detail.parameters.session_id;
  if (raw === undefined || raw === null || raw === "") {
    event.detail.parameters.session_id = 0;
    return;
  }
  var parsed = Number(raw);
  if (!Number.isNaN(parsed)) {
    event.detail.parameters.session_id = parsed;
  }
});
document.body.addEventListener("keydown", function(event) {
  if (!event.target || event.target.id !== "seer_assistant_chat_message") {
    return;
  }
  if (event.key !== "Enter" || event.shiftKey) {
    return;
  }
  event.preventDefault();
  var form = event.target.form;
  if (form) {
    form.requestSubmit();
  }
});
document.body.addEventListener("htmx:wsAfterSend", function(event) {
  if (!event.target || event.target.id !== "seer_assistant_chat_form") {
    return;
  }
  var ta = document.getElementById("seer_assistant_chat_message");
  var btn = document.getElementById("seer_assistant_chat_send");
  if (ta) {
    ta.value = "";
  }
  if (btn) {
    btn.disabled = true;
  }
});`)),
		Div(ID("seer_assistant_errors")),
		Div(
			ID("seer_assistant_transcript"),
			Class("flex flex-col gap-2 flex-1 overflow-y-auto border border-base-300 rounded-lg p-3 bg-base-200/40 min-h-[200px]"),
			Group(transcriptInner),
		),
		Div(
			ID("seer_assistant_stream"),
			Class("min-h-[1.5rem] border border-dashed border-base-300 rounded p-2 text-sm"),
		),
		FormEl(
			ID("seer_assistant_chat_form"),
			Class("flex flex-col gap-2"),
			Attr("ws-send", ""),
			Input(
				ID("seer_assistant_session_id"),
				Type("hidden"),
				Name("session_id"),
				Value(hiddenVal),
			),
			Textarea(
				ID("seer_assistant_chat_message"),
				Name("message"),
				Class("textarea textarea-bordered w-full"),
				Rows("3"),
				Placeholder("Message…"),
				Required(),
			),
			Button(
				ID("seer_assistant_chat_send"),
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
	contents, err := LoadSessionContents(ctx, db, sessionID)
	if err != nil {
		return nil, err
	}
	out := make([]Node, 0, len(contents))
	for _, c := range contents {
		inner := strings.TrimSpace(assistantGenaiContentHTML(c))
		if inner == "" {
			continue
		}
		switch assistantTranscriptTurnKind(c) {
		case "assistant":
			out = append(out, assistantBubbleAssistantHTML(inner))
		case "tool":
			out = append(out, assistantBubbleToolHTML(inner))
		default:
			out = append(out, assistantBubbleUserHTML(inner))
		}
	}
	return out, nil
}

// assistantBubble*HTML: inner is from assistantGenaiContentHTML (escaped leaves); use Raw, not Text.
func assistantBubbleUserHTML(inner string) Node {
	return Div(
		Class("chat chat-end mb-2"),
		Div(Class("chat-header text-xs opacity-70"), Text("You")),
		Div(Class("chat-bubble chat-bubble-primary"), Raw(inner)),
	)
}

func assistantBubbleAssistantHTML(inner string) Node {
	return Div(
		Class("chat chat-start mb-2"),
		Div(Class("chat-header text-xs opacity-70"), Text("Assistant")),
		Div(Class("chat-bubble chat-bubble-secondary"), Raw(inner)),
	)
}

func assistantBubbleToolHTML(inner string) Node {
	return Div(
		Class("chat chat-start mb-2"),
		Div(Class("chat-header text-xs opacity-70"), Text("Tool")),
		Div(Class("chat-bubble chat-bubble-neutral text-sm text-base-content"), Raw(inner)),
	)
}

func init() {
	registerAssistantMenuPages()
	registerAssistantChatPage()
}
