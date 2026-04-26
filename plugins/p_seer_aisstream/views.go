package p_seer_aisstream

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

var aisStreamMessageListQueryPatchers = views.QueryPatchers[AISStreamMessage]{
	{Key: "seer_aisstream.message_list.order", Value: views.QueryPatcherOrderBy[AISStreamMessage]{Order: "id DESC"}},
}

func init() {
	lago.RegistryView.Register("seer_aisstream.MapView",
		lago.GetPageView("seer_aisstream.MapPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}))

	lago.RegistryView.Register("seer_aisstream.MessageListView",
		lago.GetPageView("seer_aisstream.MessageTablePage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_aisstream.message.list", views.LayerList[AISStreamMessage]{
				Key:           getters.Static("aisStreamMessages"),
				PageSize:      getters.Static(uint(25)),
				QueryPatchers: aisStreamMessageListQueryPatchers,
			}))

	lago.RegistryView.Register("seer_aisstream.MessageDetailView",
		lago.GetPageView("seer_aisstream.MessageDetailPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_aisstream.message.detail", views.LayerDetail[AISStreamMessage]{
				Key:          getters.Static("aisStreamMessage"),
				PathParamKey: getters.Static("id"),
			}))
}
