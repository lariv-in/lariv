package p_announcements

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var announcementPreload = views.QueryPatchers[Announcement]{
	registry.Pair[string, views.QueryPatcher[Announcement]]{Key: "announcements.fk", Value: views.QueryPatcherPreload[Announcement]{Fields: []string{"Semester", "CreatedBy", "SignedBy"}}},
}

func init() {
	lago.RegistryView.Register("announcements.ListView",
		lago.GetPageView("announcements.AnnouncementTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("announcements.list", views.LayerList[Announcement]{Key: getters.Static("announcements")}))
	lago.RegistryView.Register("announcements.DetailView",
		lago.GetPageView("announcements.AnnouncementDetail").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("announcements.detail", views.LayerDetail[Announcement]{Key: getters.Static("announcement"), PathParamKey: getters.Static("id"), QueryPatchers: announcementPreload}))
	lago.RegistryView.Register("announcements.CreateView",
		lago.GetPageView("announcements.AnnouncementCreateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("announcements.create", views.LayerCreate[Announcement]{SuccessURL: lago.RoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))
	lago.RegistryView.Register("announcements.UpdateView",
		lago.GetPageView("announcements.AnnouncementUpdateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("announcements.detail", views.LayerDetail[Announcement]{Key: getters.Static("announcement"), PathParamKey: getters.Static("id"), QueryPatchers: announcementPreload}).
			WithLayer("announcements.update", views.LayerUpdate[Announcement]{Key: getters.Static("announcement"), SuccessURL: lago.RoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("announcement.ID"))})}))
	lago.RegistryView.Register("announcements.DeleteView",
		lago.GetPageView("announcements.AnnouncementDeleteForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("announcements.detail", views.LayerDetail[Announcement]{Key: getters.Static("announcement"), PathParamKey: getters.Static("id")}).
			WithLayer("announcements.delete", views.LayerDelete[Announcement]{Key: getters.Static("announcement"), SuccessURL: lago.RoutePath("announcements.DefaultRoute", nil)}))
}
