package p_seer_assistant

import (
	"context"

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
		},
	})
}

type assistantChatRoot struct {
	components.Page
}

func (e *assistantChatRoot) Build(ctx context.Context) Node {
	wsPath := AppUrl + "ws/"
	return Div(
		Class("max-w-3xl mx-auto p-4 flex flex-col gap-4 min-h-[60vh]"),
		Attr("hx-ext", "ws"),
		Attr("ws-connect", wsPath),
		Div(ID("seer_assistant_errors")),
		Div(
			ID("seer_assistant_transcript"),
			Class("flex flex-col gap-2 flex-1 overflow-y-auto border border-base-300 rounded-lg p-3 bg-base-200/40 min-h-[200px]"),
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
				Value("0"),
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

func init() {
	registerAssistantMenuPages()
	registerAssistantChatPage()
}
