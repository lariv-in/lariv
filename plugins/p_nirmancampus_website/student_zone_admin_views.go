package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

var studentZoneAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware{Roles: []string{"admin"}}

func init() {
	// --- Section views ---

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionListView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithMiddleware("student_zone_admin.section_list", views.MiddlewareList[StudentZoneSection]{
				Key: getters.Static("sections"),
				QueryPatchers: views.QueryPatchers[StudentZoneSection]{
					{Key: "student_zone_admin.order", Value: views.QueryPatcherOrderBy[StudentZoneSection]{Order: "\"order\" ASC"}},
				},
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionDetailView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithMiddleware("student_zone_admin.section_detail", views.MiddlewareDetail[StudentZoneSection]{
				Key:          getters.Static("section"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionCreateView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionCreateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithMiddleware("student_zone_admin.section_create", views.MiddlewareCreate[StudentZoneSection]{
				SuccessURL: lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionUpdateView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionUpdateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithMiddleware("student_zone_admin.section_detail", views.MiddlewareDetail[StudentZoneSection]{
				Key:          getters.Static("section"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("student_zone_admin.section_update", views.MiddlewareUpdate[StudentZoneSection]{
				Key: getters.Static("section"),
				SuccessURL: lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionDeleteView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionDeleteForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithMiddleware("student_zone_admin.section_detail", views.MiddlewareDetail[StudentZoneSection]{
				Key:          getters.Static("section"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("student_zone_admin.section_delete", views.MiddlewareDelete[StudentZoneSection]{
				Key:        getters.Static("section"),
				SuccessURL: lago.RoutePath("nirmancampus_website.StudentZoneAdminDefaultRoute", nil),
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionSelectView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionSelectionTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithMiddleware("student_zone_admin.section_select", views.MiddlewareList[StudentZoneSection]{
				Key: getters.Static("sections"),
				QueryPatchers: views.QueryPatchers[StudentZoneSection]{
					{Key: "student_zone_admin.order", Value: views.QueryPatcherOrderBy[StudentZoneSection]{Order: "\"order\" ASC"}},
				},
			}))

	// --- Item views ---

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemListView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminItemTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithMiddleware("student_zone_admin.item_list", views.MiddlewareList[StudentZoneItem]{
				Key: getters.Static("items"),
				QueryPatchers: views.QueryPatchers[StudentZoneItem]{
					{Key: "student_zone_admin.preload_section", Value: views.QueryPatcherPreload[StudentZoneItem]{Field: "StudentZoneSection"}},
					{Key: "student_zone_admin.preload_file", Value: views.QueryPatcherPreload[StudentZoneItem]{Field: "File"}},
				},
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemDetailView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminItemDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithMiddleware("student_zone_admin.item_detail", views.MiddlewareDetail[StudentZoneItem]{
				Key:          getters.Static("item"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[StudentZoneItem]{
					{Key: "student_zone_admin.preload_section", Value: views.QueryPatcherPreload[StudentZoneItem]{Field: "StudentZoneSection"}},
					{Key: "student_zone_admin.preload_file", Value: views.QueryPatcherPreload[StudentZoneItem]{Field: "File"}},
				},
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemCreateView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminItemCreateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithMiddleware("student_zone_admin.item_create", views.MiddlewareCreate[StudentZoneItem]{
				SuccessURL: lago.RoutePath("nirmancampus_website.StudentZoneAdminItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemUpdateView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminItemUpdateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithMiddleware("student_zone_admin.item_detail", views.MiddlewareDetail[StudentZoneItem]{
				Key:          getters.Static("item"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[StudentZoneItem]{
					{Key: "student_zone_admin.preload_section", Value: views.QueryPatcherPreload[StudentZoneItem]{Field: "StudentZoneSection"}},
					{Key: "student_zone_admin.preload_file", Value: views.QueryPatcherPreload[StudentZoneItem]{Field: "File"}},
				},
			}).
			WithMiddleware("student_zone_admin.item_update", views.MiddlewareUpdate[StudentZoneItem]{
				Key: getters.Static("item"),
				SuccessURL: lago.RoutePath("nirmancampus_website.StudentZoneAdminItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemDeleteView",
		lago.GetPageView("nirmancampus_website.StudentZoneAdminItemDeleteForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithMiddleware("student_zone_admin.item_detail", views.MiddlewareDetail[StudentZoneItem]{
				Key:          getters.Static("item"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("student_zone_admin.item_delete", views.MiddlewareDelete[StudentZoneItem]{
				Key:        getters.Static("item"),
				SuccessURL: lago.RoutePath("nirmancampus_website.StudentZoneAdminItemListRoute", nil),
			}))
}
