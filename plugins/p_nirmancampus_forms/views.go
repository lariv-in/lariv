package p_nirmancampus_forms

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_forms"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	forms.PublicSubmitSuccessRedirectURL = func(*forms.Form) string {
		return "/"
	}

	// Superusers are allowed by RoleAuthorizationLayer; everyone else must have role "admin".
	adminOnly := p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}
	for _, name := range []string{
		"forms.ListView",
		"forms.DetailView",
		"forms.CreateView",
		"forms.UpdateView",
		"forms.DeleteView",
		"forms.FieldCreateView",
		"forms.FieldUpdateView",
		"forms.FieldDeleteView",
		"forms.FieldMoveUpView",
		"forms.FieldMoveDownView",
		"forms.SubmissionsListView",
		"forms.SubmissionDetailView",
	} {
		viewName := name
		lago.RegistryView.Patch(viewName, func(v *views.View) *views.View {
			return v.InsertLayerAfter("users.auth", "p_nirmancampus_forms.admin", adminOnly)
		})
	}
}
