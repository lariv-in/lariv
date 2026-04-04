package sqlagent

import (
	"context"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func init() {
	registerSidebarMenu()
	registerConversationPages()
}

func registerSidebarMenu() {
	lago.RegistryPage.Register("sqlagent.SidebarMenu", &components.SidebarMenu{
		Title: getters.Static("Chats"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All chats"),
				Url:   lago.RoutePath("sqlagent.DefaultRoute", nil),
			},
		},
	})
}

func registerConversationPages() {
	lago.RegistryPage.Register("sqlagent.ConversationListPage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sqlagent.SidebarMenu"},
			&conversationSidebarList{Page: components.Page{Key: "sqlagent.SidebarList"}},
		},
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Page: components.Page{Key: "sqlagent.ListMain"},
				Children: []components.PageInterface{
					&components.FieldTitle{Getter: getters.Static("SQL Agent")},
					&components.FieldText{
						Page:    components.Page{Key: "sqlagent.ListHint"},
						Getter:  getters.Static("Select a conversation or start a new one."),
						Classes: "text-base-content/70 mb-4",
					},
					&components.FormComponent[Conversation]{
						Page:     components.Page{Key: "sqlagent.NewConversationForm"},
						OnSubmit: getters.FormSubmit(lago.RoutePath("sqlagent.ConversationCreateRoute", nil)),
						Method:   http.MethodPost,
						Title:    "",
						Subtitle: "",
						ChildrenInput: []components.PageInterface{
							&components.InputText{
								Page:   components.Page{Key: "sqlagent.NewConversationTitle"},
								Label:  "",
								Name:   "Title",
								Hidden: true,
								Getter: getters.Static("New conversation"),
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "New conversation"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("sqlagent.ConversationDetailPage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sqlagent.SidebarMenu"},
			&conversationSidebarList{Page: components.Page{Key: "sqlagent.SidebarListDetail"}},
		},
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Page: components.Page{Key: "sqlagent.DetailMain"},
				Children: []components.PageInterface{
					&components.FieldTitle{
						Page:   components.Page{Key: "sqlagent.DetailTitle"},
						Getter: getters.Key[string]("conversation.Title"),
					},
					&chatPanel{
						Page: components.Page{Key: "sqlagent.ChatPanel"},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("sqlagent.ConversationCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "sqlagent.SidebarMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Conversation]{
				OnSubmit: getters.FormSubmit(lago.RoutePath("sqlagent.ConversationCreateRoute", nil)),
				Method:   http.MethodPost,
				Title:    "New conversation",
				ChildrenInput: []components.PageInterface{
					&components.InputText{
						Name:   "Title",
						Hidden: true,
						Getter: getters.Static("New conversation"),
					},
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Create"},
				},
			},
		},
	})
}

// conversationSidebarList renders conversation links from context "conversations".
type conversationSidebarList struct {
	components.Page
}

func (e conversationSidebarList) GetKey() string     { return e.Key }
func (e conversationSidebarList) GetRoles() []string { return e.Roles }

func (e conversationSidebarList) Build(ctx context.Context) Node {
	ol, _ := ctx.Value("conversations").(components.ObjectList[Conversation])
	activeID, _ := ctx.Value(ContextKeyActiveConversationID).(uint)
	var items []Node
	if ol.Items != nil {
		for _, c := range ol.Items {
			cid := c.ID
			title := c.Title
			if title == "" {
				title = "(untitled)"
			}
			href, err := lago.RoutePath("sqlagent.ConversationDetailRoute", map[string]getters.Getter[any]{
				"conversation_id": getters.Any(getters.Static(cid)),
			})(ctx)
			if err != nil {
				logError("sqlagent: conversation detail route", err, "conversation_id", cid)
				href = "#"
			}
			activeClass := ""
			if activeID != 0 && cid == activeID {
				activeClass = "menu-active"
			}
			items = append(items, Li(
				A(Href(href), Class("truncate block "+activeClass), Text(title)),
			))
		}
	}
	return Ul(Class("menu menu-sm rounded-box w-full"), Group(items))
}

type chatPanel struct {
	components.Page
}

func (e chatPanel) GetKey() string     { return e.Key }
func (e chatPanel) GetRoles() []string { return e.Roles }

func (e chatPanel) Build(ctx context.Context) Node {
	msgs, _ := ctx.Value(ContextKeyMessages).([]ConversationMessage)
	if msgs == nil {
		msgs = []ConversationMessage{}
	}
	wsURL := ""
	if g, err := lago.RoutePath("sqlagent.WSRoute", map[string]getters.Getter[any]{
		"conversation_id": getters.Any(getters.Key[uint]("conversation.ID")),
	})(ctx); err == nil {
		wsURL = g
	}
	var bubbles Group
	for _, m := range msgs {
		htmlStr := RenderMessageBubble(m)
		if htmlStr != "" {
			bubbles = append(bubbles, Raw(htmlStr))
		}
	}
	return Div(
		Class("flex flex-col gap-2 min-h-0"),
		Div(
			Attr("hx-ext", "ws"),
			Attr("ws-connect", wsURL),
			Class("flex flex-col gap-3 min-h-[320px]"),
			Div(
				ID("sqlagent-transcript"),
				Class("flex-1 min-h-[200px] max-h-[60vh] overflow-y-auto border border-base-300 rounded-lg p-3 bg-base-100"),
				bubbles,
			),
			Form(
				Class("flex flex-col gap-2"),
				Attr("ws-send", ""),
				Textarea(
					Name("content"),
					Class("textarea textarea-bordered w-full"),
					Placeholder("Message…"),
					Attr("rows", "3"),
				),
				Button(Type("submit"), Class("btn btn-primary"), Text("Send")),
			),
		),
	)
}
