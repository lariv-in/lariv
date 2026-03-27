package p_nirmancampus_studentapplications

import (
	"net/http"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func applicationDOBFormPatcher(_ *views.View, _ *http.Request, formData map[string]any) map[string]any {
	raw, ok := formData["DOB"]
	if !ok {
		return formData
	}
	if raw == nil {
		formData["DOB"] = nil
		return formData
	}
	switch typed := raw.(type) {
	case time.Time:
		if typed.IsZero() {
			formData["DOB"] = nil
			return formData
		}
		// Store calendar date only (matches gorm type:date, like Tally.Date).
		d := time.Date(typed.Year(), typed.Month(), typed.Day(), 0, 0, 0, 0, typed.Location())
		formData["DOB"] = &d
	case *time.Time:
	default:
	}
	return formData
}

func init() {
	lago.RegistryView.Register("studentapplications.ListView",
		views.ListView[StudentApplication]("studentapplications")(
			lago.GetPageView("studentapplications.ApplicationTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")))

	lago.RegistryView.Register("studentapplications.DetailView",
		views.DetailView[StudentApplication]("studentapplication")(
			lago.GetPageView("studentapplications.ApplicationDetail")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("studentapplications.preload_photo", views.QueryPatcherPreload("Photo")).
			WithQueryPatcher("studentapplications.preload_documents", views.QueryPatcherPreload("Documents")))

	lago.RegistryView.Register("studentapplications.CreateView",
		views.CreateView[StudentApplication](lago.GetterRoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
			lago.GetPageView("studentapplications.ApplicationCreateForm")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithFormPatcher("studentapplications.form_dob", applicationDOBFormPatcher))

	lago.RegistryView.Register("studentapplications.UpdateView",
		views.DetailView[StudentApplication]("studentapplication")(
			views.UpdateView[StudentApplication](lago.GetterRoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
				lago.GetPageView("studentapplications.ApplicationUpdateForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("studentapplications.preload_photo", views.QueryPatcherPreload("Photo")).
			WithQueryPatcher("studentapplications.preload_documents", views.QueryPatcherPreload("Documents")).
			WithFormPatcher("studentapplications.form_dob", applicationDOBFormPatcher))

	lago.RegistryView.Register("studentapplications.DeleteView",
		views.DetailView[StudentApplication]("studentapplication")(
			views.DeleteView[StudentApplication](lago.GetterRoutePath("studentapplications.DefaultRoute", nil))(
				lago.GetPageView("studentapplications.ApplicationDeleteForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")))
}
