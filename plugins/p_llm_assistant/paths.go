package p_llm_assistant

import (
	"net/http"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"golang.org/x/net/websocket"
)

func init() {
	registerPluginRoute("llm_assistant.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("llm_assistant.ChatView"),
	})

	registerPluginRoute("llm_assistant.HistoryRoute", lago.Route{
		Path:    AppUrl + "history/",
		Handler: lago.NewDynamicView("llm_assistant.HistoryView"),
	})

	registerPluginRoute("llm_assistant.ChatSessionRoute", lago.Route{
		Path:    AppUrl + "c/{id}/",
		Handler: lago.NewDynamicView("llm_assistant.ChatSessionView"),
	})

	registerPluginRoute("llm_assistant.SidebarChatRoute", lago.Route{
		Path:    AppUrl + "sidebar-chat/{id}/",
		Handler: lago.NewDynamicView("llm_assistant.SidebarChatView"),
	})

	registerPluginRoute("llm_assistant.NewSessionRoute", lago.Route{
		Path:    AppUrl + "new-session/",
		Handler: p_users.RequireAuth(http.HandlerFunc(handleNewSession)),
	})

	registerPluginRoute("llm_assistant.WSRoute", lago.Route{
		Path: AppUrl + "ws/",
		Handler: p_users.RequireAuth(websocket.Server{
			Handler: assistantWebSocketConn,
		}),
	})

	registerPluginRoute("llm_assistant.SkillsListRoute", lago.Route{
		Path:    AppUrl + "skills/",
		Handler: lago.NewDynamicView("llm_assistant.SkillsListView"),
	})

	registerPluginRoute("llm_assistant.SkillsCreateRoute", lago.Route{
		Path:    AppUrl + "skills/create/",
		Handler: lago.NewDynamicView("llm_assistant.SkillsCreateView"),
	})

	registerPluginRoute("llm_assistant.SkillsDetailRoute", lago.Route{
		Path:    AppUrl + "skills/{id}/",
		Handler: lago.NewDynamicView("llm_assistant.SkillsDetailView"),
	})

	registerPluginRoute("llm_assistant.SkillsUpdateRoute", lago.Route{
		Path:    AppUrl + "skills/{id}/update/",
		Handler: lago.NewDynamicView("llm_assistant.SkillsUpdateView"),
	})

	registerPluginRoute("llm_assistant.SkillsDeleteRoute", lago.Route{
		Path:    AppUrl + "skills/{id}/delete/",
		Handler: lago.NewDynamicView("llm_assistant.SkillsDeleteView"),
	})

	registerPluginRoute("llm_assistant.SkillsExportRoute", lago.Route{
		Path:    AppUrl + "skills/{id}/export/",
		Handler: p_users.RequireAuth(http.HandlerFunc(handleSkillExport)),
	})

	registerPluginRoute("llm_assistant.SkillsImportRoute", lago.Route{
		Path:    AppUrl + "skills/import/",
		Handler: p_users.RequireAuth(http.HandlerFunc(handleSkillImportRoute)),
	})
}
