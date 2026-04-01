package p_nirmancampus_studentapplications

import (
	"net/http"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

// studentApplicationsAccessMiddleware allows only admin, Unassigned, and (via bypass) superuser.
// Student and other roles cannot use this app.
var studentApplicationsAccessMiddleware = p_users.RoleAuthorizationMiddleware([]string{"admin", roleNameUnassigned})

// studentApplicationsAdminMiddleware allows create/update/delete management (not for Unassigned except create is separate).
var studentApplicationsAdminMiddleware = p_users.RoleAuthorizationMiddleware([]string{"admin"})

func applicationCreatedByFormPatcher(_ *views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	user := r.Context().Value("$user").(p_users.User)
	id := user.ID
	formData["CreatedByID"] = &id
	return formData, formErrors
}

func applicationDOBFormPatcher(_ *views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
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
		views.ListView[StudentApplication]("studentapplications")(
			lago.GetPageView("studentapplications.ApplicationTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("studentapplications.access", studentApplicationsAccessMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("studentapplications.scope_by_role", StudentApplicationScopeByRole))

	lago.RegistryView.Register("studentapplications.DetailView",
		views.DetailView[StudentApplication]("studentapplication", "id")(
			lago.GetPageView("studentapplications.ApplicationDetail")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("studentapplications.path", views.PathMiddleware("id")).
			WithMiddleware("studentapplications.access", studentApplicationsAccessMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("studentapplications.preload_photo", views.QueryPatcherPreload("Photo")).
			WithQueryPatcher("studentapplications.preload_documents", views.QueryPatcherPreload("Documents")).
			WithQueryPatcher("studentapplications.scope_by_role", StudentApplicationScopeByRole))

	lago.RegistryView.Register("studentapplications.CreateView",
		views.CreateView[StudentApplication](lago.RoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}))(
			lago.GetPageView("studentapplications.ApplicationCreateForm")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("studentapplications.access", studentApplicationsAccessMiddleware).
			WithFormPatcher("studentapplications.form_created_by", applicationCreatedByFormPatcher).
			WithFormPatcher("studentapplications.form_dob", applicationDOBFormPatcher))

	lago.RegistryView.Register("studentapplications.UpdateView",
		views.DetailView[StudentApplication]("studentapplication", "id")(
			views.UpdateView[StudentApplication]("id", lago.RoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}))(
				lago.GetPageView("studentapplications.ApplicationUpdateForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("studentapplications.path", views.PathMiddleware("id")).
			WithMiddleware("studentapplications.access", studentApplicationsAccessMiddleware).
			WithMiddleware("studentapplications.admin_role", studentApplicationsAdminMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("studentapplications.preload_photo", views.QueryPatcherPreload("Photo")).
			WithQueryPatcher("studentapplications.preload_documents", views.QueryPatcherPreload("Documents")).
			WithQueryPatcher("studentapplications.scope_by_role", StudentApplicationScopeByRole).
			WithFormPatcher("studentapplications.form_dob", applicationDOBFormPatcher))

	lago.RegistryView.Register("studentapplications.DeleteView",
		views.DetailView[StudentApplication]("studentapplication", "id")(
			views.DeleteView[StudentApplication]("id", lago.RoutePath("studentapplications.DefaultRoute", nil))(
				lago.GetPageView("studentapplications.ApplicationDeleteForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("studentapplications.path", views.PathMiddleware("id")).
			WithMiddleware("studentapplications.access", studentApplicationsAccessMiddleware).
			WithMiddleware("studentapplications.admin_role", studentApplicationsAdminMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("studentapplications.scope_by_role", StudentApplicationScopeByRole))
}
