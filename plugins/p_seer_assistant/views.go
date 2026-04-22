package p_seer_assistant

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	sessionListPatchers := views.QueryPatchers[SeerAssistantSession]{
		{Key: "seer_assistant.session.user_scope", Value: assistantSessionUserScope{}},
		{Key: "seer_assistant.session.order", Value: views.QueryPatcherOrderBy[SeerAssistantSession]{Order: "updated_at DESC"}},
	}
	sessionDetailPatchers := views.QueryPatchers[SeerAssistantSession]{
		{Key: "seer_assistant.session.user_scope", Value: assistantSessionUserScope{}},
	}

	lago.RegistryView.Register("seer_assistant.ChatView",
		lago.GetPageView("seer_assistant.ChatPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}))

	lago.RegistryView.Register("seer_assistant.HistoryView",
		lago.GetPageView("seer_assistant.HistoryPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_assistant.session.list", views.LayerList[SeerAssistantSession]{
				Key:           getters.Static("assistantSessions"),
				QueryPatchers: sessionListPatchers,
			}))

	lago.RegistryView.Register("seer_assistant.ChatSessionView",
		lago.GetPageView("seer_assistant.ChatPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_assistant.session.detail", views.LayerDetail[SeerAssistantSession]{
				Key:           getters.Static("assistantSession"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: sessionDetailPatchers,
			}))
}
