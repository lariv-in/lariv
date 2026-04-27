package p_forums

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var forumThreadPreload = views.QueryPatchers[ForumThread]{
	registry.Pair[string, views.QueryPatcher[ForumThread]]{Key: "forums.fk", Value: views.QueryPatcherPreload[ForumThread]{Fields: []string{"Course", "Author"}}},
}

func init() {
	lago.RegistryView.Register("forums.ListView",
		lago.GetPageView("forums.ForumThreadTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("forums.list", views.LayerList[ForumThread]{Key: getters.Static("forum_threads")}))
	lago.RegistryView.Register("forums.DetailView",
		lago.GetPageView("forums.ForumThreadDetail").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("forums.detail", views.LayerDetail[ForumThread]{Key: getters.Static("forum_thread"), PathParamKey: getters.Static("id"), QueryPatchers: forumThreadPreload}))
	lago.RegistryView.Register("forums.CreateView",
		lago.GetPageView("forums.ForumThreadCreateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("forums.create", views.LayerCreate[ForumThread]{SuccessURL: lago.RoutePath("forums.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))
	lago.RegistryView.Register("forums.UpdateView",
		lago.GetPageView("forums.ForumThreadUpdateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("forums.detail", views.LayerDetail[ForumThread]{Key: getters.Static("forum_thread"), PathParamKey: getters.Static("id"), QueryPatchers: forumThreadPreload}).
			WithLayer("forums.update", views.LayerUpdate[ForumThread]{Key: getters.Static("forum_thread"), SuccessURL: lago.RoutePath("forums.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("forum_thread.ID"))})}))
	lago.RegistryView.Register("forums.DeleteView",
		lago.GetPageView("forums.ForumThreadDeleteForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("forums.detail", views.LayerDetail[ForumThread]{Key: getters.Static("forum_thread"), PathParamKey: getters.Static("id")}).
			WithLayer("forums.delete", views.LayerDelete[ForumThread]{Key: getters.Static("forum_thread"), SuccessURL: lago.RoutePath("forums.DefaultRoute", nil)}))
}
