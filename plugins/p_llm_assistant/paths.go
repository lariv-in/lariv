package p_llm_assistant

import (
	"net/http"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/plugins/p_users"
	"golang.org/x/net/websocket"
)

func init() {
	registerPluginRoute("llm_assistant.DefaultRoute", lariv.Route{
		Path:    AppUrl,
		Handler: lariv.NewDynamicView("llm_assistant.ChatView"),
	})

	registerPluginRoute("llm_assistant.HistoryRoute", lariv.Route{
		Path:    AppUrl + "history/",
		Handler: lariv.NewDynamicView("llm_assistant.HistoryView"),
	})

	registerPluginRoute("llm_assistant.ChatSessionRoute", lariv.Route{
		Path:    AppUrl + "c/{id}/",
		Handler: lariv.NewDynamicView("llm_assistant.ChatSessionView"),
	})

	registerPluginRoute("llm_assistant.SidebarChatRoute", lariv.Route{
		Path:    AppUrl + "sidebar-chat/{id}/",
		Handler: lariv.NewDynamicView("llm_assistant.SidebarChatView"),
	})

	registerPluginRoute("llm_assistant.NewSessionRoute", lariv.Route{
		Path:    AppUrl + "new-session/",
		Handler: p_users.RequireAuth(http.HandlerFunc(handleNewSession)),
	})

	registerPluginRoute("llm_assistant.WSRoute", lariv.Route{
		Path: AppUrl + "ws/",
		Handler: p_users.RequireAuth(websocket.Server{
			Handler: assistantWebSocketConn,
		}),
	})

	registerPluginRoute("llm_assistant.SkillsListRoute", lariv.Route{
		Path:    AppUrl + "skills/",
		Handler: lariv.NewDynamicView("llm_assistant.SkillsListView"),
	})

	registerPluginRoute("llm_assistant.SkillsCreateRoute", lariv.Route{
		Path:    AppUrl + "skills/create/",
		Handler: lariv.NewDynamicView("llm_assistant.SkillsCreateView"),
	})

	registerPluginRoute("llm_assistant.SkillsDetailRoute", lariv.Route{
		Path:    AppUrl + "skills/{id}/",
		Handler: lariv.NewDynamicView("llm_assistant.SkillsDetailView"),
	})

	registerPluginRoute("llm_assistant.SkillsUpdateRoute", lariv.Route{
		Path:    AppUrl + "skills/{id}/update/",
		Handler: lariv.NewDynamicView("llm_assistant.SkillsUpdateView"),
	})

	registerPluginRoute("llm_assistant.SkillsDeleteRoute", lariv.Route{
		Path:    AppUrl + "skills/{id}/delete/",
		Handler: lariv.NewDynamicView("llm_assistant.SkillsDeleteView"),
	})

	registerPluginRoute("llm_assistant.SkillsExportRoute", lariv.Route{
		Path:    AppUrl + "skills/{id}/export/",
		Handler: p_users.RequireAuth(http.HandlerFunc(handleSkillExport)),
	})

	registerPluginRoute("llm_assistant.SkillsImportRoute", lariv.Route{
		Path:    AppUrl + "skills/import/",
		Handler: p_users.RequireAuth(http.HandlerFunc(handleSkillImportRoute)),
	})
}
