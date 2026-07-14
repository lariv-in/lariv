package p_llm_assistant

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
	. "maragu.dev/gomponents/html"
)

func registerAssistantMenuPages() {
	registerPluginPage("llm_assistant.AssistantMenu", &components.SidebarMenu{
		Title: getters.Static("Assistant"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Chat"),
				Url:   lago.RoutePath("llm_assistant.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("History"),
				Url:   lago.RoutePath("llm_assistant.HistoryRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Skills"),
				Url:   lago.RoutePath("llm_assistant.SkillsListRoute", nil),
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
	rootClass := "max-w-3xl mx-auto p-4 flex flex-col gap-4 min-h-[60vh]"
	transcriptClass := "flex flex-col gap-2 flex-1 overflow-y-auto border border-base-300 rounded-lg p-3 bg-base-200/40 min-h-[200px]"
	if e.Key == "llm_assistant.SidebarChatInner" {
		rootClass = "max-w-3xl mx-auto p-0 flex flex-col gap-4 h-full overflow-hidden"
		transcriptClass = "flex flex-col gap-2 flex-1 overflow-y-auto border border-base-300 rounded-lg p-3 bg-base-200/40 min-h-0"
	}

	multiSelectUrl, _ := lago.RoutePath("filesystem.MultiSelectRoute", nil)(ctx)
	multiUploadUrl, _ := lago.RoutePath("filesystem.ChatUploadRoute", nil)(ctx)

	return Div(
		Class(rootClass),
		Attr("hx-ext", "ws"),
		Attr("ws-connect", wsPath),
		Script(Raw(`document.body.addEventListener("htmx:wsConfigSend", function(event) {
  if (!event || !event.detail || !event.detail.parameters) {
    return;
  }
  if (!event.target || event.target.id !== "llm_assistant_chat_form") {
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
  if (!event.target || event.target.id !== "llm_assistant_chat_message") {
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
  if (!event.target || event.target.id !== "llm_assistant_chat_form") {
    return;
  }
  var ta = document.getElementById("llm_assistant_chat_message");
  var btn = document.getElementById("llm_assistant_chat_send");
  if (ta) {
    ta.value = "";
  }
  if (btn) {
    btn.disabled = true;
  }
  var formEl = document.getElementById("llm_assistant_chat_form");
  if (formEl && window.Alpine) {
    var data = window.Alpine.$data(formEl);
    if (data) {
      data.items = [];
    }
  }
});
function scrollToBottom() {
  var transcript = document.getElementById("llm_assistant_transcript");
  if (transcript) {
    transcript.scrollTop = transcript.scrollHeight;
  }
}
document.addEventListener("DOMContentLoaded", scrollToBottom);
if (!window.llm_assistant_scroll_registered) {
  window.llm_assistant_scroll_registered = true;
  
  var observer = new IntersectionObserver(function(entries) {
    entries.forEach(function(entry) {
      if (entry.isIntersecting) {
        scrollToBottom();
      }
    });
  });

  var observeTranscript = function() {
    var transcript = document.getElementById("llm_assistant_transcript");
    if (transcript) {
      observer.observe(transcript);
    }
  };

  observeTranscript();

  document.body.addEventListener("htmx:oobAfterSwap", function(event) {
    if (event.detail && event.detail.target && event.detail.target.id === "llm_assistant_transcript") {
      scrollToBottom();
      observeTranscript();
    }
  });
  document.body.addEventListener("htmx:afterSwap", function(event) {
    if (event.detail && event.detail.target && (event.detail.target.id === "sidebar-chat-container" || event.detail.target.querySelector("#llm_assistant_transcript"))) {
      scrollToBottom();
      observeTranscript();
    }
  });
}`)),
		Div(ID("llm_assistant_errors")),
		Div(
			ID("llm_assistant_transcript"),
			Class(transcriptClass),
			Attr("x-init", "$nextTick(() => { $el.scrollTop = $el.scrollHeight; setTimeout(() => { $el.scrollTop = $el.scrollHeight }, 100) })"),
			Group(transcriptInner),
		),
		Div(
			ID("llm_assistant_stream"),
			Class("w-full max-w-2xl mx-auto mb-4 min-h-[1.5rem] border border-dashed border-base-300 rounded-lg p-4 text-sm"),
		),
		html.Form(
			ID("llm_assistant_chat_form"), Class("flex flex-col gap-2"), Attr("ws-send", ""),
			Attr("x-data", `{
				items: [],
				uploading: false,
				syncStore() {
					if (typeof Alpine !== 'undefined') {
						if (!Alpine.store('m2mSelections')) {
							Alpine.store('m2mSelections', {});
						}
						Alpine.store('m2mSelections')['Files'] = this.items;
					}
				},
				hasItem(value) {
					value = String(value);
					return this.items.some(item => item.Key === value);
				},
				addItem(detail) {
					const value = String(detail.value);
					if (this.hasItem(value)) return;
					const display = detail.display ? String(detail.display) : value;
					this.items = [...this.items, { Key: value, Value: display }];
					this.syncStore();
				},
				removeItem(value) {
					this.items = this.items.filter(item => item.Key !== String(value));
					this.syncStore();
				},
				eventHandler(ev) {
					if (ev.detail.name === 'Files') {
						if (!this.hasItem(ev.detail.value)) {
							this.addItem(ev.detail);
						} else {
							this.removeItem(ev.detail.value);
						}
					}
				},
				async uploadFiles(fileInput) {
					if (!fileInput.files || fileInput.files.length === 0) return;
					this.uploading = true;
					try {
						const fd = new FormData();
						for (const f of fileInput.files) { fd.append('Files', f); }
						const resp = await fetch('`+multiUploadUrl+`', {
							method: 'POST',
							headers: { 'HX-Request': 'true' },
							body: fd
						});
						const data = await resp.json();
						if (Array.isArray(data)) {
							for (const node of data) {
								this.addItem({ value: String(node.id), display: node.name });
							}
						}
					} catch(e) {
						console.error('upload failed', e);
					} finally {
						this.uploading = false;
						fileInput.value = '';
					}
				}
			}`),
			Attr("x-init", "syncStore()"),
			Attr("@fk-multi-select.window", "eventHandler($event)"),
			Input(ID("llm_assistant_session_id"), Type("hidden"), Name("session_id"), Value(hiddenVal)),

			Template(
				Attr("x-for", "item in items"),
				Attr(":key", "item.Key"),
				Input(Type("hidden"), Name("Files"), Attr(":value", "item.Key")),
			),

			Div(
				Class("flex flex-wrap gap-2"),
				Attr("x-show", "items.length > 0"),
				Template(
					Attr("x-for", "item in items"),
					Attr(":key", "item.Key"),
					Div(
						Class("flex items-center gap-1 rounded-lg bg-base-200 pl-2 pr-1 py-1 text-xs"),
						Span(Class("truncate max-w-[150px]"), Attr("x-text", "item.Value")),
						Button(
							Type("button"),
							Class("btn btn-ghost btn-square btn-xs shrink-0"),
							Attr("@click.stop", "removeItem(item.Key)"),
							components.Render(components.Icon{Name: "x-mark"}, ctx),
						),
					),
				),
			),

			Textarea(ID("llm_assistant_chat_message"), Name("message"), Class("textarea textarea-bordered w-full"), Rows("3"), Placeholder("Message…"), Required()),

			Div(
				Class("flex justify-end items-center gap-2"),
				Label(
					Class("btn btn-outline btn-square"),
					Attr(":class", "uploading ? 'loading loading-spinner' : ''"),
					Attr("title", "Upload files from device"),
					Input(
						Type("file"),
						Class("hidden"),
						Multiple(),
						Attr("@change", "uploadFiles($event.target)"),
					),
					components.Render(components.Icon{Name: "arrow-up-tray"}, ctx),
				),
				Button(
					Type("button"),
					Class("btn btn-outline btn-square"),
					Attr("hx-get", multiSelectUrl+"?target_input=Files"),
					Attr("hx-target", "body"),
					Attr("hx-swap", "beforeend"),
					Attr("hx-push-url", "false"),
					components.Render(components.Icon{Name: "paper-clip"}, ctx),
				),
				Button(ID("llm_assistant_chat_send"), Type("submit"), Class("btn btn-primary"), Text("Send")),
			),
		),
	)
}

func (e *assistantChatRoot) GetKey() string { return e.Key }

func (e *assistantChatRoot) GetRoles() []string { return e.Roles }

func registerAssistantChatPage() {
	registerPluginPage("llm_assistant.ChatPage", &components.ShellScaffold{
		Page: components.Page{Key: "llm_assistant.ChatPage"},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "llm_assistant.AssistantMenu"},
		},
		Children: []components.PageInterface{
			&assistantChatRoot{
				Page: components.Page{Key: "llm_assistant.ChatInner"},
			},
		},
	})
}

func assistantOpenSessionID(ctx context.Context) uint {
	if v := ctx.Value("assistantSession"); v != nil {
		if s, ok := v.(LlmAssistantSession); ok {
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
		inner := strings.TrimSpace(assistantGenaiContentHTML(ctx, c))
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

func assistantBubbleUserHTML(inner string) Node {
	return Div(
		Class("w-full flex flex-col items-center"),
		Div(Class("w-full max-w-2xl bg-base-300/30 border border-base-300/50 rounded-xl text-sm p-2"), Raw(inner)),
	)
}

func assistantBubbleAssistantHTML(inner string) Node {
	return Div(
		Class("w-full flex flex-col items-center"),
		Div(Class("w-full max-w-2xl text-sm"), Raw(inner)),
	)
}

func assistantBubbleToolHTML(inner string) Node {
	return Div(
		Class("w-full flex flex-col"),
		El(
			"details",
			Class("collapse text-sm w-fit"),
			El("summary", Class("text-xs text-gray-300 cursor-pointer p-0"), Text("Tool Execution")),
			Div(Class("collapse-content p-3 pt-0 overflow-x-auto"), Raw(inner)),
		),
	)
}

type historySidebarPanel struct {
	components.Page
}

func (e *historySidebarPanel) Build(ctx context.Context) Node {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return Div(Class("text-error"), Text("Error: no database context"))
	}

	var sessions []LlmAssistantSession
	if err := db.Order("updated_at desc").Find(&sessions).Error; err != nil {
		return Div(Class("text-error"), Text("Error loading sessions"))
	}

	var sessionItems []Node
	for _, s := range sessions {
		title := strings.TrimSpace(s.Title)
		if title == "" {
			title = fmt.Sprintf("Session #%d", s.ID)
		}
		sessionItems = append(sessionItems, Div(
			Class("p-3 hover:bg-base-300 rounded cursor-pointer transition border-b border-base-300 last:border-b-0 text-sm block no-underline text-base-content"),
			Attr("hx-get", fmt.Sprintf("/llm-assistant/sidebar-chat/%d/", s.ID)),
			Attr("hx-target", "#sidebar-chat-container"),
			Attr("hx-swap", "innerHTML"),
			Attr("hx-push-url", "false"),
			Attr("@click", fmt.Sprintf("activeSessionId = %d; showModal = false", s.ID)),
			Text(title),
		))
	}

	if len(sessionItems) == 0 {
		sessionItems = []Node{
			Div(Class("p-4 text-center text-sm opacity-50"), Text("No sessions found")),
		}
	}

	currentSessionID := assistantOpenSessionID(ctx)
	var initialChatContent Node = Group{}
	var activeSessionName string

	if currentSessionID != 0 {
		for _, s := range sessions {
			if s.ID == currentSessionID {
				activeSessionName = strings.TrimSpace(s.Title)
				break
			}
		}
		if activeSessionName == "" {
			activeSessionName = fmt.Sprintf("Session #%d", currentSessionID)
		}

		chatInterface := components.Render(&assistantChatRoot{
			Page: components.Page{Key: "llm_assistant.SidebarChatInner"},
		}, ctx)

		initialChatContent = Div(
			Class("flex-1 overflow-hidden min-h-0"),
			chatInterface,
		)
	} else {
		initialChatContent = Div(
			Class("flex-1 overflow-hidden min-h-0"),
			Attr("hx-push-url", "false"),
		)
	}

	xData := fmt.Sprintf(`{
		showModal: false,
		activeSessionId: $persist(0).as('llm-assistant-sidebar-active-session-id'),
		init() {
			const serverSessionId = %d;
			if (serverSessionId !== 0) {
				this.activeSessionId = serverSessionId;
			} else {
				this.$nextTick(() => {
					if (this.activeSessionId !== 0) {
						const targetEl = document.getElementById('sidebar-chat-container');
						if (targetEl) {
							htmx.ajax('GET', '/llm-assistant/sidebar-chat/' + this.activeSessionId + '/', {
								target: targetEl,
								swap: 'innerHTML',
								source: targetEl
							});
						}
					}
				});
			}
		}
	}`, currentSessionID)

	return Div(
		Attr("x-data", xData),
		Attr("@new-session-created.window", "activeSessionId = $event.detail.id; showModal = false; htmx.ajax('GET', '/llm-assistant/sidebar-chat/' + activeSessionId + '/', {target: '#sidebar-chat-container', swap: 'innerHTML', source: $el})"),
		Class("flex flex-col gap-0 p-2 h-full overflow-hidden"),
		Attr("hx-push-url", "false"),

		// Header Row: Session name on left, buttons (History, New Chat) on right
		Div(
			Class("flex justify-between items-center flex-none border-b border-base-300 pb-2 px-1"),
			Div(
				ID("session-name-container"),
				Class("text-sm font-semibold truncate max-w-[70%]"),
				Text(activeSessionName),
			),
			Div(
				Class("flex gap-1 flex-none"),
				// History Button
				Button(
					Class("btn btn-sm btn-ghost btn-circle"),
					Attr("@click", "showModal = true"),
					components.Render(components.Icon{Name: "clock"}, ctx),
				),
				// New Chat Button
				Button(
					Class("btn btn-sm btn-ghost btn-circle"),
					Attr("hx-post", "/llm-assistant/new-session/"),
					Attr("hx-swap", "none"),
					Attr("hx-push-url", "false"),
					components.Render(components.Icon{Name: "plus"}, ctx),
				),
			),
		),

		// Selected Session Name & Chat under the button (swapped dynamically)
		Div(
			ID("sidebar-chat-container"),
			Class("flex-1 flex flex-col gap-4 overflow-hidden min-h-0"),
			Attr("hx-push-url", "false"),
			initialChatContent,
		),

		// Custom Modal using standard dialog element, controlled by Alpine
		El(
			"dialog",
			Attr("x-show", "showModal"),
			Attr(":class", "showModal ? 'modal modal-open' : 'modal'"),
			Div(
				Class("modal-box bg-base-100 max-w-lg border border-base-300 p-6 relative"),
				// Close button
				Button(
					Type("button"),
					Class("btn btn-sm btn-circle btn-ghost absolute right-3 top-3"),
					Attr("@click", "showModal = false"),
					components.Render(components.Icon{Name: "x-mark"}, ctx),
				),
				// Modal Title
				H3(Class("text-lg font-bold mb-4"), Text("Conversations")),
				// Sessions List
				Div(
					ID("modal-sessions-list"),
					Class("max-h-60 overflow-y-auto flex flex-col bg-base-200 rounded border border-base-300"),
					Group(sessionItems),
				),
			),
			// Backdrop clicking closes the modal
			FormEl(
				Method("dialog"),
				Class("modal-backdrop"),
				Button(Attr("@click", "showModal = false"), Text("close")),
			),
		),
	)
}

func (e *historySidebarPanel) GetKey() string     { return e.Key }
func (e *historySidebarPanel) GetRoles() []string { return e.Roles }

// sidebarChatPage is rendered dynamically inside the sidebar container when a session is switched.
type sidebarChatPage struct {
	components.Page
}

func (e *sidebarChatPage) Build(ctx context.Context) Node {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return Div(Class("text-error"), Text("Error: no database context"))
	}
	currentSessionID := assistantOpenSessionID(ctx)
	if currentSessionID == 0 {
		return Div(Class("text-error"), Text("No session selected"))
	}
	var session LlmAssistantSession
	if err := db.First(&session, currentSessionID).Error; err != nil {
		return Div(Class("text-error"), Text("Session not found"))
	}

	title := strings.TrimSpace(session.Title)
	if title == "" {
		title = fmt.Sprintf("Session #%d", session.ID)
	}

	chatInterface := components.Render(&assistantChatRoot{
		Page: components.Page{Key: "llm_assistant.SidebarChatInner"},
	}, ctx)

	return Group{
		Div(
			ID("session-name-container"),
			Attr("hx-swap-oob", "true"),
			Class("text-sm font-semibold truncate max-w-[70%]"),
			Text(title),
		),
		Div(
			Class("flex-1 overflow-hidden min-h-0"),
			chatInterface,
		),
	}
}

func (e *sidebarChatPage) GetKey() string     { return e.Key }
func (e *sidebarChatPage) GetRoles() []string { return e.Roles }

func sidebarChatPageLookup(name string) (components.PageInterface, bool) {
	if name == "llm_assistant.SidebarChatPage" {
		return &sidebarChatPage{
			Page: components.Page{Key: "llm_assistant.SidebarChatPage"},
		}, true
	}
	return nil, false
}

func init() {
	registerAssistantMenuPages()
	registerAssistantChatPage()

	components.RegistryRightSidebar.Register("llm_assistant.history_panel", components.SidebarItem{
		Icon: "clock",
		Content: &historySidebarPanel{
			Page: components.Page{Key: "llm_assistant.history_panel"},
		},
	})
}
