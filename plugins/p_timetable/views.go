package p_timetable

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var timetableSlotPreload = views.QueryPatchers[TimetableSlot]{
	registry.Pair[string, views.QueryPatcher[TimetableSlot]]{Key: "timetable.fk", Value: views.QueryPatcherPreload[TimetableSlot]{Fields: []string{"Course", "Semester"}}},
}

func init() {
	lago.RegistryView.Register("timetable.ListView",
		lago.GetPageView("timetable.TimetableSlotTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("timetable.list", views.LayerList[TimetableSlot]{Key: getters.Static("timetable_slots"), QueryPatchers: timetableSlotPreload}))
	lago.RegistryView.Register("timetable.DetailView",
		lago.GetPageView("timetable.TimetableSlotDetail").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("timetable.detail", views.LayerDetail[TimetableSlot]{Key: getters.Static("timetable_slot"), PathParamKey: getters.Static("id"), QueryPatchers: timetableSlotPreload}))
	lago.RegistryView.Register("timetable.CreateView",
		lago.GetPageView("timetable.TimetableSlotCreateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("timetable.create", views.LayerCreate[TimetableSlot]{SuccessURL: lago.RoutePath("timetable.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))
	lago.RegistryView.Register("timetable.UpdateView",
		lago.GetPageView("timetable.TimetableSlotUpdateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("timetable.detail", views.LayerDetail[TimetableSlot]{Key: getters.Static("timetable_slot"), PathParamKey: getters.Static("id"), QueryPatchers: timetableSlotPreload}).
			WithLayer("timetable.update", views.LayerUpdate[TimetableSlot]{Key: getters.Static("timetable_slot"), SuccessURL: lago.RoutePath("timetable.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("timetable_slot.ID"))})}))
	lago.RegistryView.Register("timetable.DeleteView",
		lago.GetPageView("timetable.TimetableSlotDeleteForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("timetable.detail", views.LayerDetail[TimetableSlot]{Key: getters.Static("timetable_slot"), PathParamKey: getters.Static("id"), QueryPatchers: timetableSlotPreload}).
			WithLayer("timetable.delete", views.LayerDelete[TimetableSlot]{Key: getters.Static("timetable_slot"), SuccessURL: lago.RoutePath("timetable.DefaultRoute", nil)}))
}
