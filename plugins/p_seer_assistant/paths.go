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

	wsHandler := p_users.RequireAuth(websocketUpgradeHandler())
	_ = lago.RegistryRoute.Register("seer_assistant.WSRoute", lago.Route{
		Path:    AppUrl + "ws/",
		Handler: wsHandler,
	})
}
