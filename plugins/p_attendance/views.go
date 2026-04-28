package p_attendance

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var attendancePreload = views.QueryPatchers[AttendanceMark]{
	registry.Pair[string, views.QueryPatcher[AttendanceMark]]{Key: "attendance.fk", Value: views.QueryPatcherPreload[AttendanceMark]{Fields: []string{"Student", "Session", "Course", "Program", "Semester"}}},
}

func init() {
	lago.RegistryView.Register("attendance.ListView",
		lago.GetPageView("attendance.AttendanceMarkTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("attendance.list", views.LayerList[AttendanceMark]{Key: getters.Static("attendance_marks")}))
	lago.RegistryView.Register("attendance.DetailView",
		lago.GetPageView("attendance.AttendanceMarkDetail").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("attendance.detail", views.LayerDetail[AttendanceMark]{Key: getters.Static("attendance_mark"), PathParamKey: getters.Static("id"), QueryPatchers: attendancePreload}))
	lago.RegistryView.Register("attendance.CreateView",
		lago.GetPageView("attendance.AttendanceMarkCreateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("attendance.create", views.LayerCreate[AttendanceMark]{SuccessURL: lago.RoutePath("attendance.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))
	lago.RegistryView.Register("attendance.UpdateView",
		lago.GetPageView("attendance.AttendanceMarkUpdateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("attendance.detail", views.LayerDetail[AttendanceMark]{Key: getters.Static("attendance_mark"), PathParamKey: getters.Static("id"), QueryPatchers: attendancePreload}).
			WithLayer("attendance.update", views.LayerUpdate[AttendanceMark]{Key: getters.Static("attendance_mark"), SuccessURL: lago.RoutePath("attendance.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("attendance_mark.ID"))})}))
	lago.RegistryView.Register("attendance.DeleteView",
		lago.GetPageView("attendance.AttendanceMarkDeleteForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("attendance.detail", views.LayerDetail[AttendanceMark]{Key: getters.Static("attendance_mark"), PathParamKey: getters.Static("id"), QueryPatchers: attendancePreload}).
			WithLayer("attendance.delete", views.LayerDelete[AttendanceMark]{Key: getters.Static("attendance_mark"), SuccessURL: lago.RoutePath("attendance.DefaultRoute", nil)}))
}
