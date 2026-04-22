package p_seer_assistant

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func init() {
	_ = lago.RegistryRoute.Register("seer_assistant.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_assistant.ChatView"),
	})

	_ = lago.RegistryRoute.Register("seer_assistant.HistoryRoute", lago.Route{
		Path:    AppUrl + "history/",
		Handler: lago.NewDynamicView("seer_assistant.HistoryView"),
	})

	_ = lago.RegistryRoute.Register("seer_assistant.ChatSessionRoute", lago.Route{
		Path:    AppUrl + "c/{id}/",
		Handler: lago.NewDynamicView("seer_assistant.ChatSessionView"),
	})

	wsHandler := p_users.RequireAuth(websocketUpgradeHandler())
	_ = lago.RegistryRoute.Register("seer_assistant.WSRoute", lago.Route{
		Path:    AppUrl + "ws/",
		Handler: wsHandler,
	})
}
