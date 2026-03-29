package forms

import "github.com/lariv-in/lago/lago"

func init() {
	_ = lago.RegistryRoute.Register("forms.DefaultRoute", lago.Route{
		Path:    AppURL,
		Handler: lago.NewDynamicView("forms.ListView"),
	})
	_ = lago.RegistryRoute.Register("forms.CreateRoute", lago.Route{
		Path:    AppURL + "create/",
		Handler: lago.NewDynamicView("forms.CreateView"),
	})
	_ = lago.RegistryRoute.Register("forms.DetailRoute", lago.Route{
		Path:    AppURL + "{form_id}/",
		Handler: lago.NewDynamicView("forms.DetailView"),
	})
	_ = lago.RegistryRoute.Register("forms.UpdateRoute", lago.Route{
		Path:    AppURL + "{form_id}/edit/",
		Handler: lago.NewDynamicView("forms.UpdateView"),
	})
	_ = lago.RegistryRoute.Register("forms.DeleteRoute", lago.Route{
		Path:    AppURL + "{form_id}/delete/",
		Handler: lago.NewDynamicView("forms.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("forms.FieldCreateRoute", lago.Route{
		Path:    AppURL + "{form_id}/fields/create/",
		Handler: lago.NewDynamicView("forms.FieldCreateView"),
	})
	// Field routes use {form_id} for the form and {id} for the field; submission detail uses {form_id} + {id} for the submission.
	_ = lago.RegistryRoute.Register("forms.FieldUpdateRoute", lago.Route{
		Path:    AppURL + "{form_id}/fields/{id}/edit/",
		Handler: lago.NewDynamicView("forms.FieldUpdateView"),
	})
	_ = lago.RegistryRoute.Register("forms.FieldDeleteRoute", lago.Route{
		Path:    AppURL + "{form_id}/fields/{id}/delete/",
		Handler: lago.NewDynamicView("forms.FieldDeleteView"),
	})
	_ = lago.RegistryRoute.Register("forms.FieldMoveUpRoute", lago.Route{
		Path:    AppURL + "{form_id}/fields/{id}/move-up/",
		Handler: lago.NewDynamicView("forms.FieldMoveUpView"),
	})
	_ = lago.RegistryRoute.Register("forms.FieldMoveDownRoute", lago.Route{
		Path:    AppURL + "{form_id}/fields/{id}/move-down/",
		Handler: lago.NewDynamicView("forms.FieldMoveDownView"),
	})

	_ = lago.RegistryRoute.Register("forms.SubmissionsListRoute", lago.Route{
		Path:    AppURL + "{form_id}/submissions/",
		Handler: lago.NewDynamicView("forms.SubmissionsListView"),
	})
	// Nested under form_id so this does not conflict with /forms/{form_id}/fields/ (e.g. /forms/submissions/fields/).
	_ = lago.RegistryRoute.Register("forms.SubmissionDetailRoute", lago.Route{
		Path:    AppURL + "{form_id}/submissions/{id}/",
		Handler: lago.NewDynamicView("forms.SubmissionDetailView"),
	})

	// Literal prefix "public/p" avoids conflicting with /forms/{id}/submissions/ (e.g. /forms/p/submissions/).
	_ = lago.RegistryRoute.Register("forms.PublicFormRoute", lago.Route{
		Path:    "/forms/public/p/{slug}/",
		Handler: lago.NewDynamicView("forms.PublicSubmitView"),
	})
}
