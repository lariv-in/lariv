package p_llm_assistant

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func init() {
	sessionListPatchers := views.QueryPatchers[LlmAssistantSession]{
		{Key: "llm_assistant.session.user_scope", Value: assistantSessionUserScope{}},
		{Key: "llm_assistant.session.order", Value: views.QueryPatcherOrderBy[LlmAssistantSession]{Order: "updated_at DESC"}},
	}
	sessionDetailPatchers := views.QueryPatchers[LlmAssistantSession]{
		{Key: "llm_assistant.session.user_scope", Value: assistantSessionUserScope{}},
	}

	registerPluginView("llm_assistant.ChatView",
		lago.GetPageView("llm_assistant.ChatPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}))

	registerPluginView("llm_assistant.HistoryView",
		lago.GetPageView("llm_assistant.HistoryPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("llm_assistant.session.list", views.LayerList[LlmAssistantSession]{
				Key:           getters.Static("assistantSessions"),
				QueryPatchers: sessionListPatchers,
			}))

	registerPluginView("llm_assistant.ChatSessionView",
		lago.GetPageView("llm_assistant.ChatPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("llm_assistant.session.detail", views.LayerDetail[LlmAssistantSession]{
				Key:           getters.Static("assistantSession"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: sessionDetailPatchers,
			}))

	registerPluginView("llm_assistant.SidebarChatView",
		&views.View{
			PageName:   "llm_assistant.SidebarChatPage",
			PageLookup: sidebarChatPageLookup,
			Layers: []registry.Pair[string, views.Layer]{
				{Key: "p_users.auth", Value: p_users.AuthenticationLayer{}},
				{Key: "llm_assistant.session.detail", Value: views.LayerDetail[LlmAssistantSession]{
					Key:           getters.Static("assistantSession"),
					PathParamKey:  getters.Static("id"),
					QueryPatchers: sessionDetailPatchers,
				}},
			},
		})

	registerPluginView("llm_assistant.SkillsListView",
		lago.GetPageView("llm_assistant.SkillsListPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("llm_assistant.skills.list", views.LayerList[Skill]{
				Key: getters.Static("skills"),
			}))

	registerPluginView("llm_assistant.SkillsCreateView",
		lago.GetPageView("llm_assistant.SkillsCreatePage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("llm_assistant.skills.create", views.LayerCreate[Skill]{
				SuccessURL: lago.RoutePath("llm_assistant.SkillsDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	registerPluginView("llm_assistant.SkillsDetailView",
		lago.GetPageView("llm_assistant.SkillsDetailPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("llm_assistant.skills.detail", views.LayerDetail[Skill]{
				Key:          getters.Static("skill"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Skill]{
					{Key: "llm_assistant.skills.preload", Value: views.QueryPatcherPreload[Skill]{Fields: []string{"Files"}}},
				},
			}))

	registerPluginView("llm_assistant.SkillsUpdateView",
		lago.GetPageView("llm_assistant.SkillsUpdatePage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("llm_assistant.skills.detail", views.LayerDetail[Skill]{
				Key:          getters.Static("skill"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Skill]{
					{Key: "llm_assistant.skills.preload", Value: views.QueryPatcherPreload[Skill]{Fields: []string{"Files"}}},
				},
			}).
			WithLayer("llm_assistant.skills.update", views.LayerUpdate[Skill]{
				Key: getters.Static("skill"),
				SuccessURL: lago.RoutePath("llm_assistant.SkillsDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("skill.ID")),
				}),
			}))

	registerPluginView("llm_assistant.SkillsDeleteView",
		lago.GetPageView("llm_assistant.SkillsDeletePage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("llm_assistant.skills.detail", views.LayerDetail[Skill]{
				Key:          getters.Static("skill"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("llm_assistant.skills.delete", views.LayerDelete[Skill]{
				Key:        getters.Static("skill"),
				SuccessURL: lago.RoutePath("llm_assistant.SkillsListRoute", nil),
			}))

	registerPluginView("llm_assistant.SkillsImportView",
		lago.GetPageView("llm_assistant.SkillsImportPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}))
}

func handleNewSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("$user").(p_users.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		http.Error(w, "No database connection", http.StatusInternalServerError)
		return
	}
	session, err := CreateSession(ctx, db, user.ID)
	if err != nil {
		http.Error(w, "Could not create session", http.StatusInternalServerError)
		return
	}

	var sessions []LlmAssistantSession
	if err := db.Order("updated_at desc").Find(&sessions).Error; err != nil {
		http.Error(w, "Could not load sessions", http.StatusInternalServerError)
		return
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

	oobList := Div(
		ID("modal-sessions-list"),
		Attr("hx-swap-oob", "innerHTML"),
		Group(sessionItems),
	)

	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"new-session-created": {"id": %d}}`, session.ID))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_ = oobList.Render(w)
}
