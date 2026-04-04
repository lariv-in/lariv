package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

var studentZoneAdminRoleLayer = p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}

func init() {
	// --- Section views ---

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionListView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("student_zone_admin.role", studentZoneAdminRoleLayer).
			WithLayer("student_zone_admin.section_list", views.LayerList[StudentZoneSection]{
				Key: getters.Static("sections"),
				QueryPatchers: views.QueryPatchers[StudentZoneSection]{
					{Key: "student_zone_admin.order", Value: views.QueryPatcherOrderBy[StudentZoneSection]{Order: "\"order\" ASC"}},
				},
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionDetailView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("student_zone_admin.role", studentZoneAdminRoleLayer).
			WithLayer("student_zone_admin.section_detail", views.LayerDetail[StudentZoneSection]{
				Key:          getters.Static("section"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionCreateView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("student_zone_admin.role", studentZoneAdminRoleLayer).
			WithLayer("student_zone_admin.section_create", views.LayerCreate[StudentZoneSection]{
				SuccessURL: lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionUpdateView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("student_zone_admin.role", studentZoneAdminRoleLayer).
			WithLayer("student_zone_admin.section_detail", views.LayerDetail[StudentZoneSection]{
				Key:          getters.Static("section"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("student_zone_admin.section_update", views.LayerUpdate[StudentZoneSection]{
				Key: getters.Static("section"),
				SuccessURL: lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("section.ID")),
				}),
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionDeleteView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("student_zone_admin.role", studentZoneAdminRoleLayer).
			WithLayer("student_zone_admin.section_detail", views.LayerDetail[StudentZoneSection]{
				Key:          getters.Static("section"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("student_zone_admin.section_delete", views.LayerDelete[StudentZoneSection]{
				Key:        getters.Static("section"),
				SuccessURL: lago.RoutePath("nirmancampus_website.StudentZoneAdminDefaultRoute", nil),
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionSelectView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("student_zone_admin.role", studentZoneAdminRoleLayer).
			WithLayer("student_zone_admin.section_select", views.LayerList[StudentZoneSection]{
				Key: getters.Static("sections"),
				QueryPatchers: views.QueryPatchers[StudentZoneSection]{
					{Key: "student_zone_admin.order", Value: views.QueryPatcherOrderBy[StudentZoneSection]{Order: "\"order\" ASC"}},
				},
			}))

	// --- Item views ---

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemListView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminItemTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("student_zone_admin.role", studentZoneAdminRoleLayer).
			WithLayer("student_zone_admin.item_list", views.LayerList[StudentZoneItem]{
				Key: getters.Static("items"),
				QueryPatchers: views.QueryPatchers[StudentZoneItem]{
					{Key: "student_zone_admin.preload_section", Value: views.QueryPatcherPreload[StudentZoneItem]{Field: "StudentZoneSection"}},
					{Key: "student_zone_admin.preload_file", Value: views.QueryPatcherPreload[StudentZoneItem]{Field: "File"}},
				},
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemDetailView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminItemDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("student_zone_admin.role", studentZoneAdminRoleLayer).
			WithLayer("student_zone_admin.item_detail", views.LayerDetail[StudentZoneItem]{
				Key:          getters.Static("item"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[StudentZoneItem]{
					{Key: "student_zone_admin.preload_section", Value: views.QueryPatcherPreload[StudentZoneItem]{Field: "StudentZoneSection"}},
					{Key: "student_zone_admin.preload_file", Value: views.QueryPatcherPreload[StudentZoneItem]{Field: "File"}},
				},
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemCreateView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminItemCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("student_zone_admin.role", studentZoneAdminRoleLayer).
			WithLayer("student_zone_admin.item_create", views.LayerCreate[StudentZoneItem]{
				SuccessURL: lago.RoutePath("nirmancampus_website.StudentZoneAdminItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemUpdateView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminItemUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("student_zone_admin.role", studentZoneAdminRoleLayer).
			WithLayer("student_zone_admin.item_detail", views.LayerDetail[StudentZoneItem]{
				Key:          getters.Static("item"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[StudentZoneItem]{
					{Key: "student_zone_admin.preload_section", Value: views.QueryPatcherPreload[StudentZoneItem]{Field: "StudentZoneSection"}},
					{Key: "student_zone_admin.preload_file", Value: views.QueryPatcherPreload[StudentZoneItem]{Field: "File"}},
				},
			}).
			WithLayer("student_zone_admin.item_update", views.LayerUpdate[StudentZoneItem]{
				Key: getters.Static("item"),
				SuccessURL: lago.RoutePath("nirmancampus_website.StudentZoneAdminItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("item.ID")),
				}),
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemDeleteView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminItemDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("student_zone_admin.role", studentZoneAdminRoleLayer).
			WithLayer("student_zone_admin.item_detail", views.LayerDetail[StudentZoneItem]{
				Key:          getters.Static("item"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("student_zone_admin.item_delete", views.LayerDelete[StudentZoneItem]{
				Key:        getters.Static("item"),
				SuccessURL: lago.RoutePath("nirmancampus_website.StudentZoneAdminItemListRoute", nil),
			}))
}
