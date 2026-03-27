package forms

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// queryPatcherScopeByParentFormID scopes ListView[FormSubmission] on /forms/{form_id}/submissions/.
// QueryPatchers also run for other handlers; skip unless the query targets FormSubmission
// (e.g. Form has no form_id column).
func queryPatcherScopeByParentFormID() views.QueryPatcher {
	return func(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
		if !queryModelHasFormIDScope(query) {
			return query
		}
		s := r.PathValue("form_id")
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil || u == 0 {
			slog.Error("forms: invalid parent form id in path", "error", err, "raw", s)
			return query.Where("1 = 0")
		}
		return query.Where("form_id = ?", uint(u))
	}
}

func queryModelHasFormIDScope(query *gorm.DB) bool {
	m := query.Statement.Model
	if m == nil {
		return false
	}
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Name() {
	case "FormSubmission":
		return true
	default:
		return false
	}
}

func queryModelIsForm(query *gorm.DB) bool {
	m := query.Statement.Model
	if m == nil {
		return false
	}
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Name() == "Form"
}

// queryPatcherSubmissionDetailPreloads preloads fields for submission detail (nested DetailView[FormSubmission])
// and for the outer DetailView[Form] on the same route.
func queryPatcherSubmissionDetailPreloads() views.QueryPatcher {
	return func(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
		if queryModelHasFormIDScope(query) {
			return query.Preload("Form.FormFields", func(db *gorm.DB) *gorm.DB {
				return db.Order("sort_order ASC, id ASC")
			})
		}
		if queryModelIsForm(query) {
			return query.Preload("FormFields", func(db *gorm.DB) *gorm.DB {
				return db.Order("sort_order ASC, id ASC")
			})
		}
		return query
	}
}

// queryPatcherFormFieldScopeByFormPath scopes FormField load/update to PathValue("form_id") (route /forms/{form_id}/fields/{id}/edit/).
func queryPatcherFormFieldScopeByFormPath() views.QueryPatcher {
	return func(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
		s := r.PathValue("form_id")
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil || u == 0 {
			slog.Error("forms: invalid form_id in field path", "error", err, "raw", s)
			return query.Where("1 = 0")
		}
		return query.Where("form_id = ?", uint(u))
	}
}

// queryPatcherSubmissionScopeByFormPath scopes FormSubmission detail to PathValue("form_id") (route /forms/{form_id}/submissions/{id}/).
// Skips other models (e.g. outer DetailView[Form]) which share the same view's query patchers.
func queryPatcherSubmissionScopeByFormPath() views.QueryPatcher {
	return func(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
		if !queryModelHasFormIDScope(query) {
			return query
		}
		s := r.PathValue("form_id")
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil || u == 0 {
			slog.Error("forms: invalid form_id in submission detail path", "error", err, "raw", s)
			return query.Where("1 = 0")
		}
		return query.Where("form_id = ?", uint(u))
	}
}

// queryPatcherPreloadFormFieldsOrdered loads FormFields in display order for form detail.
func queryPatcherPreloadFormFieldsOrdered() views.QueryPatcher {
	return func(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
		return query.Preload("FormFields", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		})
	}
}

// formPatcherCoerceFormFieldFormID converts FormID from hidden text inputs (string) to uint for PopulateFromMap.
func formPatcherCoerceFormFieldFormID() views.FormPatcher {
	return func(v *views.View, r *http.Request, formData map[string]any) map[string]any {
		raw, ok := formData["FormID"]
		if !ok || raw == nil {
			return formData
		}
		switch x := raw.(type) {
		case uint:
			return formData
		case string:
			if u, err := strconv.ParseUint(strings.TrimSpace(x), 10, 64); err == nil {
				formData["FormID"] = uint(u)
			}
		case int:
			formData["FormID"] = uint(x)
		case int64:
			formData["FormID"] = uint(x)
		case float64:
			formData["FormID"] = uint(x)
		}
		return formData
	}
}

// formPatcherNextSortOrderOnFieldCreate sets SortOrder to one greater than the current max for this form.
func formPatcherNextSortOrderOnFieldCreate() views.FormPatcher {
	return func(v *views.View, r *http.Request, formData map[string]any) map[string]any {
		var formID uint
		switch x := formData["FormID"].(type) {
		case uint:
			formID = x
		case int:
			formID = uint(x)
		case float64:
			formID = uint(x)
		case string:
			u, err := strconv.ParseUint(x, 10, 64)
			if err == nil {
				formID = uint(u)
			}
		}
		if formID == 0 {
			return formData
		}
		db, ok := r.Context().Value("$db").(*gorm.DB)
		if !ok || db == nil {
			return formData
		}
		var rows []FormField
		db.Where("form_id = ?", formID).Order("sort_order DESC").Limit(1).Find(&rows)
		next := 0
		if len(rows) > 0 {
			next = rows[0].SortOrder + 1
		}
		formData["SortOrder"] = next
		return formData
	}
}

// formPatcherFieldNameFromLabel sets Name from Label (HTML-safe slug) so admins only edit the label.
func formPatcherFieldNameFromLabel() views.FormPatcher {
	return func(v *views.View, r *http.Request, formData map[string]any) map[string]any {
		var label string
		switch x := formData["Label"].(type) {
		case string:
			label = x
		default:
			if x != nil {
				label = fmt.Sprint(x)
			}
		}
		formData["Name"] = getters.LabelToHTMLName(label)
		return formData
	}
}

// formPatcherClearOptionsUnlessSelect drops Options when FieldType is not select (the field may
// still be posted while hidden in the DOM).
func formPatcherClearOptionsUnlessSelect() views.FormPatcher {
	return func(v *views.View, r *http.Request, formData map[string]any) map[string]any {
		ft, _ := formData["FieldType"].(string)
		if ft != FieldTypeSelect {
			formData["Options"] = "[]"
		}
		return formData
	}
}

// ensureUniqueFormSlug picks a slug starting from base, appending -2, -3, … if needed.
func ensureUniqueFormSlug(db *gorm.DB, base string, excludeID uint) string {
	if base == "" {
		base = "form"
	}
	if len([]rune(base)) > 160 {
		runes := []rune(base)[:160]
		base = strings.TrimRight(string(runes), "-")
		if base == "" {
			base = "form"
		}
	}
	for i := 0; i < 1000; i++ {
		candidate := base
		if i > 0 {
			suf := fmt.Sprintf("-%d", i+1)
			prefix := base
			pr := []rune(prefix)
			for len(string(pr))+len(suf) > 160 && len(pr) > 0 {
				pr = pr[:len(pr)-1]
			}
			prefix = strings.TrimRight(string(pr), "-")
			if prefix == "" {
				prefix = "form"
			}
			candidate = prefix + suf
		}
		var count int64
		q := db.Model(&Form{}).Where("slug = ?", candidate)
		if excludeID > 0 {
			q = q.Where("id <> ?", excludeID)
		}
		if err := q.Count(&count).Error; err != nil {
			slog.Error("forms: slug uniqueness check", "error", err)
			return candidate
		}
		if count == 0 {
			return candidate
		}
	}
	return base + "-1"
}

// formPatcherSlugFromTitle sets Slug from Title (URL-safe, unique in the database).
func formPatcherSlugFromTitle() views.FormPatcher {
	return func(v *views.View, r *http.Request, formData map[string]any) map[string]any {
		var title string
		switch x := formData["Title"].(type) {
		case string:
			title = strings.TrimSpace(x)
		default:
			if x != nil {
				title = strings.TrimSpace(fmt.Sprint(x))
			}
		}
		base := getters.TitleToFormSlug(title)
		db, ok := r.Context().Value("$db").(*gorm.DB)
		if !ok || db == nil {
			formData["Slug"] = base
			return formData
		}
		var excludeID uint
		if s := r.PathValue("form_id"); s != "" {
			if u, err := strconv.ParseUint(s, 10, 64); err == nil {
				excludeID = uint(u)
			}
		}
		formData["Slug"] = ensureUniqueFormSlug(db, base, excludeID)
		return formData
	}
}

func init() {
	auth := "users.auth"

	lago.RegistryView.Register("forms.ListView",
		views.ListView[Form]("forms")(
			lago.GetPageView("forms.FormTable"),
		).
			WithMiddleware(auth, p_users.AuthenticationMiddleware).
			WithQueryPatcher("forms.order", views.QueryPatcherOrderBy("title ASC")))

	lago.RegistryView.Register("forms.DetailView",
		views.DetailView[Form]("form", "form_id")(
			lago.GetPageView("forms.FormDetail"),
		).
			WithMiddleware(auth, p_users.AuthenticationMiddleware).
			WithRenderMiddleware("forms.fields_object_list", AttachFormFieldsObjectListContext).
			WithQueryPatcher("forms.detail_preload_fields", queryPatcherPreloadFormFieldsOrdered()))

	lago.RegistryView.Register("forms.CreateView",
		views.CreateView[Form](
			lago.GetterRoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
				"form_id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("forms.FormCreateForm"),
		).
			WithMiddleware(auth, p_users.AuthenticationMiddleware).
			WithFormPatcher("forms.slug_from_title", formPatcherSlugFromTitle()))

	lago.RegistryView.Register("forms.UpdateView",
		views.DetailView[Form]("form", "form_id")(
			views.UpdateView[Form]("form_id",
				lago.GetterRoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
					"form_id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("forms.FormUpdateForm"),
			),
		).
			WithMiddleware(auth, p_users.AuthenticationMiddleware).
			WithFormPatcher("forms.slug_from_title", formPatcherSlugFromTitle()))

	lago.RegistryView.Register("forms.DeleteView",
		views.DetailView[Form]("form", "form_id")(
			views.DeleteView[Form]("form_id", lago.GetterRoutePath("forms.DefaultRoute", nil))(
				lago.GetPageView("forms.FormDeleteForm"),
			),
		).
			WithMiddleware(auth, p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("forms.FieldCreateView",
		views.DetailView[Form]("form", "form_id")(
			views.CreateView[FormField](
				// CreateView sets $id; PathMiddleware sets $path (form_id, id).
				lago.GetterRoutePath("forms.FieldUpdateRoute", map[string]getters.Getter[any]{
					"form_id": getters.GetterAny(getters.GetterParseUint(getters.GetterKey[string]("$path.form_id"))),
					"id":      getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("forms.FieldCreateForm"),
			),
		).
			WithMiddleware(auth, p_users.AuthenticationMiddleware).
			WithMiddleware("forms.path_params", views.PathMiddleware("form_id", "id")).
			WithMiddleware("forms.form_parent_fields_ctx", AttachFormForParentFieldsPath).
			WithFormPatcher("forms.coerce_form_field_form_id", formPatcherCoerceFormFieldFormID()).
			WithFormPatcher("forms.field_next_sort_order", formPatcherNextSortOrderOnFieldCreate()).
			WithFormPatcher("forms.field_name_from_label", formPatcherFieldNameFromLabel()).
			WithFormPatcher("forms.clear_options_unless_select", formPatcherClearOptionsUnlessSelect()))

	lago.RegistryView.Register("forms.FieldUpdateView",
		views.DetailView[FormField]("form_field", "id")(
			views.UpdateView[FormField]("id", lago.GetterRoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
				"form_id": getters.GetterAny(getters.GetterKey[uint]("form_field.FormID")),
			}))(
				lago.GetPageView("forms.FieldUpdateForm"),
			),
		).
			WithMiddleware(auth, p_users.AuthenticationMiddleware).
			WithMiddleware("forms.path_params", views.PathMiddleware("form_id", "id")).
			WithQueryPatcher("forms.field_form_scope", queryPatcherFormFieldScopeByFormPath()).
			WithFormPatcher("forms.coerce_form_field_form_id", formPatcherCoerceFormFieldFormID()).
			WithFormPatcher("forms.field_name_from_label", formPatcherFieldNameFromLabel()).
			WithFormPatcher("forms.clear_options_unless_select", formPatcherClearOptionsUnlessSelect()))

	lago.RegistryView.Register("forms.FieldDeleteView",
		views.DetailView[FormField]("form_field", "id")(
			views.DeleteView[FormField]("id", lago.GetterRoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
				"form_id": getters.GetterAny(getters.GetterParseUint(getters.GetterKey[string]("$path.form_id"))),
			}))(
				lago.GetPageView("forms.FieldDeleteForm"),
			),
		).
			WithMiddleware(auth, p_users.AuthenticationMiddleware).
			WithMiddleware("forms.path_params", views.PathMiddleware("form_id", "id")).
			WithQueryPatcher("forms.field_form_scope", queryPatcherFormFieldScopeByFormPath()))

	lago.RegistryView.Register("forms.FieldMoveUpView", fieldMovePostView(true))
	lago.RegistryView.Register("forms.FieldMoveDownView", fieldMovePostView(false))

	lago.RegistryView.Register("forms.SubmissionsListView",
		views.ListView[FormSubmission]("form_submissions")(
			lago.GetPageView("forms.SubmissionTable"),
		).
			WithMiddleware(auth, p_users.AuthenticationMiddleware).
			WithMiddleware("forms.path_params", views.PathMiddleware("form_id")).
			WithMiddleware("forms.form_parent_fields_ctx", AttachFormForParentFieldsPath).
			WithQueryPatcher("forms.submission_scope", queryPatcherScopeByParentFormID()))

	lago.RegistryView.Register("forms.SubmissionDetailView",
		views.DetailView[Form]("form", "form_id")(
			views.DetailView[FormSubmission]("form_submission", "id")(
				lago.GetPageView("forms.SubmissionDetail"),
			).
				WithMiddleware(auth, p_users.AuthenticationMiddleware).
				WithQueryPatcher("forms.submission_form_scope", queryPatcherSubmissionScopeByFormPath()).
				WithQueryPatcher("forms.submission_detail_preloads", queryPatcherSubmissionDetailPreloads()),
		),
	)

	lago.RegistryView.Register("forms.PublicSubmitView", publicSubmitView())
}

func redirectAfterFieldMove(w http.ResponseWriter, r *http.Request, formID uint) {
	ctx := r.Context()
	u, err := lago.GetterRoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
		"form_id": getters.GetterAny(getters.GetterStatic(formID)),
	})(ctx)
	if err != nil || u == "" {
		http.Redirect(w, r, AppURL, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, u, http.StatusSeeOther)
}

func fieldMovePostView(moveUp bool) *views.View {
	view := &views.View{
		PageName: "forms.FormDetail",
		PageLookup: func(name string) (components.PageInterface, bool) {
			return lago.RegistryPage.Get(name)
		},
		Handlers: map[string]func(*views.View) http.Handler{
			http.MethodGet: func(_ *views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				})
			},
			http.MethodPost: func(_ *views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					formIDStr := r.PathValue("form_id")
					idStr := r.PathValue("id")
					fieldID64, err := strconv.ParseUint(idStr, 10, 64)
					if err != nil || fieldID64 == 0 {
						http.Error(w, "Invalid field ID", http.StatusBadRequest)
						return
					}
					formID64, err2 := strconv.ParseUint(formIDStr, 10, 64)
					if err2 != nil || formID64 == 0 {
						http.Error(w, "Invalid form ID", http.StatusBadRequest)
						return
					}
					db := r.Context().Value("$db").(*gorm.DB)
					var ff FormField
					if err := db.First(&ff, fieldID64).Error; err != nil {
						http.NotFound(w, r)
						return
					}
					if fmt.Sprint(ff.FormID) != formIDStr {
						http.NotFound(w, r)
						return
					}
					formID := uint(formID64)
					if err := reorderFormField(db, formID, uint(fieldID64), moveUp); err != nil {
						slog.Error("forms: reorder field", "error", err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					redirectAfterFieldMove(w, r, formID)
				})
			},
		},
	}
	return view.
		WithMiddleware("users.auth", p_users.AuthenticationMiddleware)
}

func publicSubmitView() *views.View {
	view := &views.View{
		PageName: "forms.PublicSubmitPage",
		PageLookup: func(name string) (components.PageInterface, bool) {
			return lago.RegistryPage.Get(name)
		},
		Handlers: map[string]func(*views.View) http.Handler{
			http.MethodGet: func(inner *views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					slug := r.PathValue("slug")
					db := r.Context().Value("$db").(*gorm.DB)
					var form Form
					if err := db.Preload("FormFields", func(db *gorm.DB) *gorm.DB {
						return db.Order("sort_order ASC, id ASC")
					}).Where("slug = ?", slug).First(&form).Error; err != nil {
						http.NotFound(w, r)
						return
					}
					ctx := r.Context()
					ctx = context.WithValue(ctx, ContextKeyPublicLoadedForm, &form)
					q := map[string]any{}
					for k, vv := range r.URL.Query() {
						if len(vv) > 0 {
							q[k] = vv[0]
						}
					}
					ctx = context.WithValue(ctx, "$get", q)
					inner.RenderPage(w, r.WithContext(ctx))
				})
			},
			http.MethodPost: func(inner *views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					slug := r.PathValue("slug")
					db := r.Context().Value("$db").(*gorm.DB)
					var form Form
					if err := db.Preload("FormFields", func(db *gorm.DB) *gorm.DB {
						return db.Order("sort_order ASC, id ASC")
					}).Where("slug = ?", slug).First(&form).Error; err != nil {
						http.NotFound(w, r)
						return
					}
					ctx := r.Context()
					ctx = context.WithValue(ctx, ContextKeyPublicLoadedForm, &form)
					q := map[string]any{}
					for k, vv := range r.URL.Query() {
						if len(vv) > 0 {
							q[k] = vv[0]
						}
					}
					ctx = context.WithValue(ctx, "$get", q)
					r = r.WithContext(ctx)

					values, fieldErrors, err := inner.ParseForm(w, r)
					if err != nil {
						return
					}
					if inner.HasErrors(fieldErrors) {
						inner.RenderWithErrors(w, r, fieldErrors, values)
						return
					}
					raw, err := json.Marshal(values)
					if err != nil {
						slog.Error("forms: marshal answers", "error", err)
						fieldErrors["_form"] = err
						inner.RenderWithErrors(w, r, fieldErrors, values)
						return
					}
					sub := FormSubmission{FormID: form.ID, Answers: datatypes.JSON(raw)}
					if err := db.Create(&sub).Error; err != nil {
						slog.Error("forms: create submission", "error", err)
						fieldErrors["_form"] = err
						inner.RenderWithErrors(w, r, fieldErrors, values)
						return
					}
					loc := ThankYouRedirectURL(&form)
					if PublicSubmitSuccessRedirectURL != nil {
						if u := PublicSubmitSuccessRedirectURL(&form); u != "" {
							loc = u
						}
					}
					http.Redirect(w, r, loc, http.StatusSeeOther)
				})
			},
		},
	}
	return view
}
