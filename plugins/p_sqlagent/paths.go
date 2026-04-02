package sqlagent

import "github.com/lariv-in/lago/lago"

const AppURL = "/sqlagent/"

func init() {
	_ = lago.RegistryRoute.Register("sqlagent.DefaultRoute", lago.Route{
		Path:    AppURL,
		Handler: lago.NewDynamicView("sqlagent.ListView"),
	})
	_ = lago.RegistryRoute.Register("sqlagent.ConversationCreateRoute", lago.Route{
		Path:    AppURL + "conversations/create/",
		Handler: lago.NewDynamicView("sqlagent.ConversationCreateView"),
	})
	_ = lago.RegistryRoute.Register("sqlagent.ConversationDetailRoute", lago.Route{
		Path:    AppURL + "{conversation_id}/",
		Handler: lago.NewDynamicView("sqlagent.ConversationDetailView"),
	})
	_ = lago.RegistryRoute.Register("sqlagent.WSRoute", lago.Route{
		Path:    AppURL + "{conversation_id}/ws/",
		Handler: wsHTTPHandler(),
	})
}
