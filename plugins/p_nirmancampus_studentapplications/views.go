package p_nirmancampus_studentapplications

import (
	"net/http"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

// studentApplicationsAccessMiddleware allows only admin, Unassigned, and (via bypass) superuser.
// Student and other roles cannot use this app.
var studentApplicationsAccessMiddleware = p_users.RoleAuthorizationMiddleware{Roles: []string{"admin", roleNameUnassigned}}

// studentApplicationsAdminMiddleware allows create/update/delete management (not for Unassigned except create is separate).
var studentApplicationsAdminMiddleware = p_users.RoleAuthorizationMiddleware{Roles: []string{"admin"}}

type applicationCreatedByFormPatcher struct{}

func (applicationCreatedByFormPatcher) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	user := r.Context().Value("$user").(p_users.User)
	id := user.ID
	formData["CreatedByID"] = &id
	return formData, formErrors
}

type applicationDOBFormPatcher struct{}

func (applicationDOBFormPatcher) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	raw, ok := formData["DOB"]
	if !ok {
		return formData, formErrors
	}
	if raw == nil {
		formData["DOB"] = nil
		return formData, formErrors
	}
	switch typed := raw.(type) {
	case time.Time:
		if typed.IsZero() {
			formData["DOB"] = nil
			return formData, formErrors
		}
		// Store calendar date only (matches gorm type:date, like Tally.Date).
		d := time.Date(typed.Year(), typed.Month(), typed.Day(), 0, 0, 0, 0, typed.Location())
		formData["DOB"] = &d
	case *time.Time:
	default:
	}
	return formData, formErrors
}

func init() {
	lago.RegistryView.Register("studentapplications.ListView",
		lago.GetPageView("studentapplications.ApplicationTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("studentapplications.access", studentApplicationsAccessMiddleware).
			WithMiddleware("studentapplications.list", views.MiddlewareList[StudentApplication]{
				Key: getters.Static("studentapplications"),
				QueryPatchers: views.QueryPatchers[StudentApplication]{
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_program", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.scope_by_role", Value: StudentApplicationScopeByRole},
				},
			}))

	lago.RegistryView.Register("studentapplications.DetailView",
		lago.GetPageView("studentapplications.ApplicationDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("studentapplications.path", views.PathMiddleware{Names: []string{"id"}}).
			WithMiddleware("studentapplications.access", studentApplicationsAccessMiddleware).
			WithMiddleware("studentapplications.detail", views.MiddlewareDetail[StudentApplication]{
				Key:          getters.Static("studentapplication"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[StudentApplication]{
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_program", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_photo", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Photo"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_documents", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Documents"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.scope_by_role", Value: StudentApplicationScopeByRole},
				},
			}))

	lago.RegistryView.Register("studentapplications.CreateView",
		lago.GetPageView("studentapplications.ApplicationCreateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("studentapplications.access", studentApplicationsAccessMiddleware).
			WithMiddleware("studentapplications.create", views.MiddlewareCreate[StudentApplication]{
				SuccessURL: lago.RoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}),
				FormPatchers: views.FormPatchers{
					registry.Pair[string, views.FormPatcher]{Key: "studentapplications.form_created_by", Value: applicationCreatedByFormPatcher{}},
					registry.Pair[string, views.FormPatcher]{Key: "studentapplications.form_dob", Value: applicationDOBFormPatcher{}},
				},
			}))

	lago.RegistryView.Register("studentapplications.UpdateView",
		lago.GetPageView("studentapplications.ApplicationUpdateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("studentapplications.path", views.PathMiddleware{Names: []string{"id"}}).
			WithMiddleware("studentapplications.access", studentApplicationsAccessMiddleware).
			WithMiddleware("studentapplications.admin_role", studentApplicationsAdminMiddleware).
			WithMiddleware("studentapplications.detail", views.MiddlewareDetail[StudentApplication]{
				Key:          getters.Static("studentapplication"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[StudentApplication]{
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_program", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_photo", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Photo"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_documents", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Documents"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.scope_by_role", Value: StudentApplicationScopeByRole},
				},
			}).
			WithMiddleware("studentapplications.update", views.MiddlewareUpdate[StudentApplication]{
				Key:        getters.Static("studentapplication"),
				SuccessURL: lago.RoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}),
				QueryPatchers: views.QueryPatchers[StudentApplication]{
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.scope_by_role", Value: StudentApplicationScopeByRole},
				},
				FormPatchers: views.FormPatchers{
					registry.Pair[string, views.FormPatcher]{Key: "studentapplications.form_dob", Value: applicationDOBFormPatcher{}},
				},
			}))

	lago.RegistryView.Register("studentapplications.DeleteView",
		lago.GetPageView("studentapplications.ApplicationDeleteForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("studentapplications.path", views.PathMiddleware{Names: []string{"id"}}).
			WithMiddleware("studentapplications.access", studentApplicationsAccessMiddleware).
			WithMiddleware("studentapplications.admin_role", studentApplicationsAdminMiddleware).
			WithMiddleware("studentapplications.detail", views.MiddlewareDetail[StudentApplication]{
				Key:          getters.Static("studentapplication"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[StudentApplication]{
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_program", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.scope_by_role", Value: StudentApplicationScopeByRole},
				},
			}).
			WithMiddleware("studentapplications.delete", views.MiddlewareDelete[StudentApplication]{
				Key:        getters.Static("studentapplication"),
				SuccessURL: lago.RoutePath("studentapplications.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[StudentApplication]{
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.scope_by_role", Value: StudentApplicationScopeByRole},
				},
			}))
}
