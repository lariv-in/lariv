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

// studentApplicationsAccessLayer allows only admin, Unassigned, and (via bypass) superuser.
// Student and other roles cannot use this app.
var studentApplicationsAccessLayer = p_users.RoleAuthorizationLayer{Roles: []string{"admin", roleNameUnassigned}}

// studentApplicationsAdminLayer allows create/update/delete management (not for Unassigned except create is separate).
var studentApplicationsAdminLayer = p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}

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
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("studentapplications.access", studentApplicationsAccessLayer).
			WithLayer("studentapplications.list", views.LayerList[StudentApplication]{
				Key: getters.Static("studentapplications"),
				QueryPatchers: views.QueryPatchers[StudentApplication]{
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_program", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.scope_by_role", Value: StudentApplicationScopeByRole},
				},
			}))

	lago.RegistryView.Register("studentapplications.DetailView",
		lago.GetPageView("studentapplications.ApplicationDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("studentapplications.path", views.PathLayer{Names: []string{"id"}}).
			WithLayer("studentapplications.access", studentApplicationsAccessLayer).
			WithLayer("studentapplications.detail", views.LayerDetail[StudentApplication]{
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
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("studentapplications.access", studentApplicationsAccessLayer).
			WithLayer("studentapplications.create", views.LayerCreate[StudentApplication]{
				SuccessURL: lago.RoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}),
				FormPatchers: views.FormPatchers{
					registry.Pair[string, views.FormPatcher]{Key: "studentapplications.form_created_by", Value: applicationCreatedByFormPatcher{}},
					registry.Pair[string, views.FormPatcher]{Key: "studentapplications.form_dob", Value: applicationDOBFormPatcher{}},
				},
			}))

	lago.RegistryView.Register("studentapplications.UpdateView",
		lago.GetPageView("studentapplications.ApplicationUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("studentapplications.path", views.PathLayer{Names: []string{"id"}}).
			WithLayer("studentapplications.access", studentApplicationsAccessLayer).
			WithLayer("studentapplications.admin_role", studentApplicationsAdminLayer).
			WithLayer("studentapplications.detail", views.LayerDetail[StudentApplication]{
				Key:          getters.Static("studentapplication"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[StudentApplication]{
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_program", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_photo", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Photo"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_documents", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Documents"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.scope_by_role", Value: StudentApplicationScopeByRole},
				},
			}).
			WithLayer("studentapplications.update", views.LayerUpdate[StudentApplication]{
				Key:        getters.Static("studentapplication"),
				SuccessURL: lago.RoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("studentapplication.ID"))}),
				QueryPatchers: views.QueryPatchers[StudentApplication]{
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.scope_by_role", Value: StudentApplicationScopeByRole},
				},
				FormPatchers: views.FormPatchers{
					registry.Pair[string, views.FormPatcher]{Key: "studentapplications.form_dob", Value: applicationDOBFormPatcher{}},
				},
			}))

	lago.RegistryView.Register("studentapplications.DeleteView",
		lago.GetPageView("studentapplications.ApplicationDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("studentapplications.path", views.PathLayer{Names: []string{"id"}}).
			WithLayer("studentapplications.access", studentApplicationsAccessLayer).
			WithLayer("studentapplications.admin_role", studentApplicationsAdminLayer).
			WithLayer("studentapplications.detail", views.LayerDetail[StudentApplication]{
				Key:          getters.Static("studentapplication"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[StudentApplication]{
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.preload_program", Value: views.QueryPatcherPreload[StudentApplication]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.scope_by_role", Value: StudentApplicationScopeByRole},
				},
			}).
			WithLayer("studentapplications.delete", views.LayerDelete[StudentApplication]{
				Key:        getters.Static("studentapplication"),
				SuccessURL: lago.RoutePath("studentapplications.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[StudentApplication]{
					registry.Pair[string, views.QueryPatcher[StudentApplication]]{Key: "studentapplications.scope_by_role", Value: StudentApplicationScopeByRole},
				},
			}))
}
