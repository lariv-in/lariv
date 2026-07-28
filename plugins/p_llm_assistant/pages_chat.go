package p_llm_assistant

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func registerAssistantMenuPages() {
	registerPluginPage("llm_assistant.AssistantMenu", &components.SidebarMenu{
		Title: getters.Static("Assistant"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lariv.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Chat"),
				Url:   lariv.RoutePath("llm_assistant.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("History"),
				Url:   lariv.RoutePath("llm_assistant.HistoryRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Skills"),
				Url:   lariv.RoutePath("llm_assistant.SkillsListRoute", nil),
			},
		},
	})
}

type assistantChatRoot struct {
	components.Page
}

const assistantChatScript = `document.body.addEventListener("htmx:wsConfigSend", function(event) {
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
}`

func (e *assistantChatRoot) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	sid := assistantOpenSessionID(ctx)
	wsPath := AppUrl + "ws/"
	if sid != 0 {
		wsPath = fmt.Sprintf("%s?session_id=%d", wsPath, sid)
	}
	hiddenVal := "0"
	if sid != 0 {
		hiddenVal = fmt.Sprintf("%d", sid)
	}
	var transcript template.HTML
	if sid != 0 {
		html, err := assistantTranscriptHTML(ctx, sid)
		if err != nil {
			transcript = template.HTML(`<div class="text-error text-sm">Could not load chat history</div>`)
		} else {
			transcript = html
		}
	}
	rootClass := "max-w-3xl mx-auto p-4 flex flex-col gap-4 min-h-[60vh]"
	transcriptClass := "flex flex-col gap-2 flex-1 overflow-y-auto border border-base-300 rounded-lg p-3 bg-base-200/40 min-h-[200px]"
	if e.Key == "llm_assistant.SidebarChatInner" {
		rootClass = "max-w-3xl mx-auto p-0 flex flex-col gap-4 h-full overflow-hidden"
		transcriptClass = "flex flex-col gap-2 flex-1 overflow-y-auto border border-base-300 rounded-lg p-3 bg-base-200/40 min-h-0"
	}

	multiSelectUrl, _ := lariv.RoutePath("filesystem.MultiSelectRoute", nil)(ctx)
	multiUploadUrl, _ := lariv.RoutePath("filesystem.ChatUploadRoute", nil)(ctx)

	iconXMark, err := components.RenderHTML(components.Icon{Name: "x-mark"}, cat, ctx)
	if err != nil {
		return err
	}
	iconUpload, err := components.RenderHTML(components.Icon{Name: "arrow-up-tray"}, cat, ctx)
	if err != nil {
		return err
	}
	iconClip, err := components.RenderHTML(components.Icon{Name: "paper-clip"}, cat, ctx)
	if err != nil {
		return err
	}

	formXData := `{
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
						const resp = await fetch('` + multiUploadUrl + `', {
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
			}`

	return executeTemplate(w, "assistant_chat_root", struct {
		RootClass       string
		WSPath          string
		ChatScript      template.JS
		TranscriptClass string
		Transcript      template.HTML
		FormXData       string
		HiddenVal       string
		IconXMark       template.HTML
		IconUpload      template.HTML
		IconClip        template.HTML
		MultiSelectURL  string
	}{
		RootClass:       rootClass,
		WSPath:          wsPath,
		ChatScript:      template.JS(assistantChatScript),
		TranscriptClass: transcriptClass,
		Transcript:      transcript,
		FormXData:       formXData,
		HiddenVal:       hiddenVal,
		IconXMark:       iconXMark,
		IconUpload:      iconUpload,
		IconClip:        iconClip,
		MultiSelectURL:  multiSelectUrl,
	})
}

func (e *assistantChatRoot) GetKey() string { return e.Key }

func (e *assistantChatRoot) GetRoles() []string { return e.Roles }

func registerAssistantChatPage() {
	registerPluginPage("llm_assistant.ChatPage", &components.ShellScaffold{
		Page: components.Page{Key: "llm_assistant.ChatPage"},
		Sidebar: []components.PageInterface{
			lariv.DynamicPage{Name: "llm_assistant.AssistantMenu"},
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

func assistantTranscriptHTML(ctx context.Context, sessionID uint) (template.HTML, error) {
	if sessionID == 0 {
		return "", nil
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return "", err
	}
	contents, err := LoadSessionContents(ctx, db, sessionID)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, c := range contents {
		inner := strings.TrimSpace(assistantGenaiContentHTML(ctx, c))
		if inner == "" {
			continue
		}
		switch assistantTranscriptTurnKind(c) {
		case "assistant":
			b.WriteString(assistantBubbleAssistantHTML(inner))
		case "tool":
			b.WriteString(assistantBubbleToolHTML(inner))
		default:
			b.WriteString(assistantBubbleUserHTML(inner))
		}
	}
	return template.HTML(b.String()), nil
}

func assistantBubbleUserHTML(inner string) string {
	return `<div class="w-full flex flex-col items-center"><div class="w-full max-w-2xl bg-base-300/30 border border-base-300/50 rounded-xl text-sm p-2">` + inner + `</div></div>`
}

func assistantBubbleAssistantHTML(inner string) string {
	return `<div class="w-full flex flex-col items-center"><div class="w-full max-w-2xl text-sm">` + inner + `</div></div>`
}

func assistantBubbleToolHTML(inner string) string {
	return `<div class="w-full flex flex-col"><details class="collapse text-sm w-fit"><summary class="text-xs text-gray-300 cursor-pointer p-0">Tool Execution</summary><div class="collapse-content p-3 pt-0 overflow-x-auto">` + inner + `</div></details></div>`
}

func renderSessionItemsHTML(sessions []LlmAssistantSession) (template.HTML, error) {
	if len(sessions) == 0 {
		var b bytes.Buffer
		if err := executeTemplate(&b, "session_list_empty", nil); err != nil {
			return "", err
		}
		return template.HTML(b.String()), nil
	}
	var b bytes.Buffer
	for _, s := range sessions {
		title := strings.TrimSpace(s.Title)
		if title == "" {
			title = fmt.Sprintf("Session #%d", s.ID)
		}
		if err := executeTemplate(&b, "session_list_item", struct {
			ID    uint
			Title string
		}{ID: s.ID, Title: title}); err != nil {
			return "", err
		}
	}
	return template.HTML(b.String()), nil
}

type historySidebarPanel struct {
	components.Page
}

func (e *historySidebarPanel) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		_, writeErr := io.WriteString(w, `<div class="text-error">Error: no database context</div>`)
		return writeErr
	}

	var sessions []LlmAssistantSession
	if err := db.Order("updated_at desc").Find(&sessions).Error; err != nil {
		_, writeErr := io.WriteString(w, `<div class="text-error">Error loading sessions</div>`)
		return writeErr
	}

	sessionItems, err := renderSessionItemsHTML(sessions)
	if err != nil {
		return err
	}

	currentSessionID := assistantOpenSessionID(ctx)
	var initialChat template.HTML
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

		chatHTML, err := components.RenderHTML(&assistantChatRoot{
			Page: components.Page{Key: "llm_assistant.SidebarChatInner"},
		}, cat, ctx)
		if err != nil {
			return err
		}
		initialChat = template.HTML(`<div class="flex-1 overflow-hidden min-h-0">` + string(chatHTML) + `</div>`)
	} else {
		initialChat = template.HTML(`<div class="flex-1 overflow-hidden min-h-0" hx-push-url="false"></div>`)
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

	iconClock, err := components.RenderHTML(components.Icon{Name: "clock"}, cat, ctx)
	if err != nil {
		return err
	}
	iconPlus, err := components.RenderHTML(components.Icon{Name: "plus"}, cat, ctx)
	if err != nil {
		return err
	}
	iconXMark, err := components.RenderHTML(components.Icon{Name: "x-mark"}, cat, ctx)
	if err != nil {
		return err
	}

	return executeTemplate(w, "history_sidebar_panel", struct {
		XData             string
		ActiveSessionName string
		IconClock         template.HTML
		IconPlus          template.HTML
		IconXMark         template.HTML
		InitialChat       template.HTML
		SessionItems      template.HTML
	}{
		XData:             xData,
		ActiveSessionName: activeSessionName,
		IconClock:         iconClock,
		IconPlus:          iconPlus,
		IconXMark:         iconXMark,
		InitialChat:       initialChat,
		SessionItems:      sessionItems,
	})
}

func (e *historySidebarPanel) GetKey() string     { return e.Key }
func (e *historySidebarPanel) GetRoles() []string { return e.Roles }

// sidebarChatPage is rendered dynamically inside the sidebar container when a session is switched.
type sidebarChatPage struct {
	components.Page
}

func (e *sidebarChatPage) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		_, writeErr := io.WriteString(w, `<div class="text-error">Error: no database context</div>`)
		return writeErr
	}
	currentSessionID := assistantOpenSessionID(ctx)
	if currentSessionID == 0 {
		_, writeErr := io.WriteString(w, `<div class="text-error">No session selected</div>`)
		return writeErr
	}
	var session LlmAssistantSession
	if err := db.First(&session, currentSessionID).Error; err != nil {
		_, writeErr := io.WriteString(w, `<div class="text-error">Session not found</div>`)
		return writeErr
	}

	title := strings.TrimSpace(session.Title)
	if title == "" {
		title = fmt.Sprintf("Session #%d", session.ID)
	}

	chatHTML, err := components.RenderHTML(&assistantChatRoot{
		Page: components.Page{Key: "llm_assistant.SidebarChatInner"},
	}, cat, ctx)
	if err != nil {
		return err
	}

	return executeTemplate(w, "sidebar_chat_page", struct {
		Title string
		Chat  template.HTML
	}{
		Title: title,
		Chat:  chatHTML,
	})
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

func pluginRightSidebar() lariv.PluginFeatures[components.SidebarItem] {
	return lariv.PluginFeatures[components.SidebarItem]{
		Entries: []registry.Pair[string, components.SidebarItem]{
			{Key: "llm_assistant.history_panel", Value: components.SidebarItem{
				Icon: "clock",
				Content: &historySidebarPanel{
					Page: components.Page{Key: "llm_assistant.history_panel"},
				},
			}},
		},
	}
}

func init() {
	registerAssistantMenuPages()
	registerAssistantChatPage()
}
