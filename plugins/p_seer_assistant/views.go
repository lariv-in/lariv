package p_seer_assistant

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func init() {
	lago.RegistryView.Register("seer_assistant.ChatView",
		lago.GetPageView("seer_assistant.ChatPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}))
}
