package forms

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// --- Query patchers (typed) ---

type formSubmissionScopeByFormID struct{}

func (formSubmissionScopeByFormID) Patch(_ views.View, r *http.Request, q gorm.ChainInterface[FormSubmission]) gorm.ChainInterface[FormSubmission] {
	s := r.PathValue("form_id")
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil || u == 0 {
		slog.Error("forms: invalid parent form id in path", "error", err, "raw", s)
		return q.Where("1 = 0")
	}
	return q.Where("form_id = ?", uint(u))
}

type formDetailPreloadFields struct{}

func (formDetailPreloadFields) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[Form]) gorm.ChainInterface[Form] {
	return q.Preload("FormFields", func(p gorm.PreloadBuilder) error {
		p.Order("sort_order ASC, id ASC")
		return nil
	})
}

type formFieldScopeByFormPath struct{}

func (formFieldScopeByFormPath) Patch(_ views.View, r *http.Request, q gorm.ChainInterface[FormField]) gorm.ChainInterface[FormField] {
	s := r.PathValue("form_id")
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil || u == 0 {
		slog.Error("forms: invalid form_id in field path", "error", err, "raw", s)
		return q.Where("1 = 0")
	}
	return q.Where("form_id = ?", uint(u))
}

type formSubmissionScopeByFormPath struct{}

func (formSubmissionScopeByFormPath) Patch(_ views.View, r *http.Request, q gorm.ChainInterface[FormSubmission]) gorm.ChainInterface[FormSubmission] {
	s := r.PathValue("form_id")
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil || u == 0 {
		slog.Error("forms: invalid form_id in submission detail path", "error", err, "raw", s)
		return q.Where("1 = 0")
	}
	return q.Where("form_id = ?", uint(u))
}

type formSubmissionDetailPreloadInner struct{}

func (formSubmissionDetailPreloadInner) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[FormSubmission]) gorm.ChainInterface[FormSubmission] {
	return q.Preload("Form", nil).Preload("Form.FormFields", func(p gorm.PreloadBuilder) error {
		p.Order("sort_order ASC, id ASC")
		return nil
	})
}

// --- Form patchers ---

type coerceFormFieldFormID struct{}

func (coerceFormFieldFormID) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	raw, ok := formData["FormID"]
	if !ok || raw == nil {
		return formData, formErrors
	}
	switch x := raw.(type) {
	case uint:
		return formData, formErrors
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
	return formData, formErrors
}

type nextSortOrderOnFieldCreate struct{}

func (nextSortOrderOnFieldCreate) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
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
		return formData, formErrors
	}
	db, ok := r.Context().Value("$db").(*gorm.DB)
	if !ok || db == nil {
		return formData, formErrors
	}
	rows, _ := gorm.G[FormField](db).Where("form_id = ?", formID).Order("sort_order DESC").Limit(1).Find(r.Context())
	next := 0
	if len(rows) > 0 {
		next = rows[0].SortOrder + 1
	}
	formData["SortOrder"] = next
	return formData, formErrors
}

type fieldNameFromLabel struct{}

func (fieldNameFromLabel) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
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
	return formData, formErrors
}

type clearOptionsUnlessSelect struct{}

func (clearOptionsUnlessSelect) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	ft, _ := formData["FieldType"].(string)
	if ft != "select" {
		formData["Options"] = "[]"
	}
	return formData, formErrors
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
	for i := range 1000 {
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

type slugFromTitle struct{}

func (slugFromTitle) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
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
		return formData, formErrors
	}
	var excludeID uint
	if s := r.PathValue("form_id"); s != "" {
		if u, err := strconv.ParseUint(s, 10, 64); err == nil {
			excludeID = uint(u)
		}
	}
	formData["Slug"] = ensureUniqueFormSlug(db, base, excludeID)
	return formData, formErrors
}

var formFieldCreatePatchers = views.FormPatchers{
	{Key: "forms.coerce_form_field_form_id", Value: coerceFormFieldFormID{}},
	{Key: "forms.field_next_sort_order", Value: nextSortOrderOnFieldCreate{}},
	{Key: "forms.field_name_from_label", Value: fieldNameFromLabel{}},
	{Key: "forms.clear_options_unless_select", Value: clearOptionsUnlessSelect{}},
}

var formFieldUpdatePatchers = views.FormPatchers{
	{Key: "forms.coerce_form_field_form_id", Value: coerceFormFieldFormID{}},
	{Key: "forms.field_name_from_label", Value: fieldNameFromLabel{}},
	{Key: "forms.clear_options_unless_select", Value: clearOptionsUnlessSelect{}},
}

func init() {
	auth := "users.auth"

	lago.RegistryView.Register("forms.ListView",
		lago.GetPageView("forms.FormTable").
			WithLayer(auth, p_users.AuthenticationLayer{}).
			WithLayer("forms.list", views.LayerList[Form]{
				Key: getters.Static("forms"),
				QueryPatchers: views.QueryPatchers[Form]{
					{Key: "forms.order", Value: views.QueryPatcherOrderBy[Form]{Order: "title ASC"}},
				},
			}))

	lago.RegistryView.Register("forms.DetailView",
		lago.GetPageView("forms.FormDetail").
			WithLayer(auth, p_users.AuthenticationLayer{}).
			WithLayer("forms.detail", views.LayerDetail[Form]{
				Key:          getters.Static("form"),
				PathParamKey: getters.Static("form_id"),
				QueryPatchers: views.QueryPatchers[Form]{
					{Key: "forms.detail_preload_fields", Value: formDetailPreloadFields{}},
				},
			}).
			WithLayer("forms.fields_object_list", AttachFormFieldsObjectListContext{}))

	lago.RegistryView.Register("forms.CreateView",
		lago.GetPageView("forms.FormCreateForm").
			WithLayer(auth, p_users.AuthenticationLayer{}).
			WithLayer("forms.create", views.LayerCreate[Form]{
				SuccessURL: lago.RoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "forms.slug_from_title", Value: slugFromTitle{}},
				},
			}))

	lago.RegistryView.Register("forms.UpdateView",
		lago.GetPageView("forms.FormUpdateForm").
			WithLayer(auth, p_users.AuthenticationLayer{}).
			WithLayer("forms.detail", views.LayerDetail[Form]{
				Key:          getters.Static("form"),
				PathParamKey: getters.Static("form_id"),
			}).
			WithLayer("forms.update", views.LayerUpdate[Form]{
				Key: getters.Static("form"),
				SuccessURL: lago.RoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "forms.slug_from_title", Value: slugFromTitle{}},
				},
			}))

	lago.RegistryView.Register("forms.DeleteView",
		lago.GetPageView("forms.FormDeleteForm").
			WithLayer(auth, p_users.AuthenticationLayer{}).
			WithLayer("forms.detail", views.LayerDetail[Form]{
				Key:          getters.Static("form"),
				PathParamKey: getters.Static("form_id"),
			}).
			WithLayer("forms.delete", views.LayerDelete[Form]{
				Key:        getters.Static("form"),
				SuccessURL: lago.RoutePath("forms.DefaultRoute", nil),
			}))

	lago.RegistryView.Register("forms.FieldCreateView",
		lago.GetPageView("forms.FieldCreateForm").
			WithLayer(auth, p_users.AuthenticationLayer{}).
			WithLayer("forms.path_params", views.PathLayer{Names: []string{"form_id", "id"}}).
			WithLayer("forms.detail_form", views.LayerDetail[Form]{
				Key:          getters.Static("form"),
				PathParamKey: getters.Static("form_id"),
			}).
			WithLayer("forms.field_create", views.LayerCreate[FormField]{
				SuccessURL: lago.RoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.ParseUint(getters.Key[string]("$path.form_id"))),
				}),
				FormPatchers: formFieldCreatePatchers,
			}))

	lago.RegistryView.Register("forms.FieldUpdateView",
		lago.GetPageView("forms.FieldUpdateForm").
			WithLayer(auth, p_users.AuthenticationLayer{}).
			WithLayer("forms.path_params", views.PathLayer{Names: []string{"form_id", "id"}}).
			WithLayer("forms.field_detail", views.LayerDetail[FormField]{
				Key:          getters.Static("form_field"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[FormField]{
					{Key: "forms.field_form_scope", Value: formFieldScopeByFormPath{}},
				},
			}).
			WithLayer("forms.field_update", views.LayerUpdate[FormField]{
				Key: getters.Static("form_field"),
				SuccessURL: lago.RoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form_field.FormID")),
				}),
				FormPatchers: formFieldUpdatePatchers,
			}))

	lago.RegistryView.Register("forms.FieldDeleteView",
		lago.GetPageView("forms.FieldDeleteForm").
			WithLayer(auth, p_users.AuthenticationLayer{}).
			WithLayer("forms.path_params", views.PathLayer{Names: []string{"form_id", "id"}}).
			WithLayer("forms.field_detail", views.LayerDetail[FormField]{
				Key:          getters.Static("form_field"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[FormField]{
					{Key: "forms.field_form_scope", Value: formFieldScopeByFormPath{}},
				},
			}).
			WithLayer("forms.field_delete", views.LayerDelete[FormField]{
				Key: getters.Static("form_field"),
				SuccessURL: lago.RoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.ParseUint(getters.Key[string]("$path.form_id"))),
				}),
			}))

	lago.RegistryView.Register("forms.FieldMoveUpView", fieldMovePostView(true))
	lago.RegistryView.Register("forms.FieldMoveDownView", fieldMovePostView(false))

	lago.RegistryView.Register("forms.SubmissionsListView",
		lago.GetPageView("forms.SubmissionTable").
			WithLayer(auth, p_users.AuthenticationLayer{}).
			WithLayer("forms.path_params", views.PathLayer{Names: []string{"form_id"}}).
			WithLayer("forms.form_parent_fields_ctx", AttachFormForParentFieldsPath{}).
			WithLayer("forms.submissions_list", views.LayerList[FormSubmission]{
				Key: getters.Static("form_submissions"),
				QueryPatchers: views.QueryPatchers[FormSubmission]{
					{Key: "forms.submission_scope", Value: formSubmissionScopeByFormID{}},
				},
			}))

	lago.RegistryView.Register("forms.SubmissionDetailView",
		lago.GetPageView("forms.SubmissionDetail").
			WithLayer(auth, p_users.AuthenticationLayer{}).
			WithLayer("forms.submission_outer", views.LayerDetail[Form]{
				Key:          getters.Static("form"),
				PathParamKey: getters.Static("form_id"),
				QueryPatchers: views.QueryPatchers[Form]{
					{Key: "forms.submission_outer_preloads", Value: formDetailPreloadFields{}},
				},
			}).
			WithLayer("forms.submission_inner", views.LayerDetail[FormSubmission]{
				Key:          getters.Static("form_submission"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[FormSubmission]{
					{Key: "forms.submission_form_scope", Value: formSubmissionScopeByFormPath{}},
					{Key: "forms.submission_detail_preloads", Value: formSubmissionDetailPreloadInner{}},
				},
			}))

	lago.RegistryView.Register("forms.PublicSubmitView", publicSubmitView())
}

func redirectAfterFieldMove(w http.ResponseWriter, r *http.Request, formID uint) {
	ctx := r.Context()
	u, err := lago.RoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
		"form_id": getters.Any(getters.Static(formID)),
	})(ctx)
	if err != nil || u == "" {
		views.HtmxRedirect(w, r, AppURL, http.StatusSeeOther)
		return
	}
	views.HtmxRedirect(w, r, u, http.StatusSeeOther)
}

func fieldMoveHandler(moveUp bool) func(*views.View) http.Handler {
	return func(_ *views.View) http.Handler {
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
			ff, err := gorm.G[FormField](db).Where("id = ?", fieldID64).First(r.Context())
			if err != nil {
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
	}
}

func fieldMovePostView(moveUp bool) *views.View {
	return lago.GetPageView("forms.FormDetail").
		WithLayer("users.auth", p_users.AuthenticationLayer{}).
		WithLayer("forms.path_params", views.PathLayer{Names: []string{"form_id", "id"}}).
		WithLayer("forms.field_move_get", views.MethodLayer{
			Method: http.MethodGet,
			Handler: func(_ *views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				})
			},
		}).
		WithLayer("forms.field_move_post", views.MethodLayer{
			Method:  http.MethodPost,
			Handler: fieldMoveHandler(moveUp),
		})
}

func publicSubmitView() *views.View {
	return lago.GetPageView("forms.PublicSubmitPage").
		WithLayer("forms.public_get", views.MethodLayer{
			Method: http.MethodGet,
			Handler: func(inner *views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					slug := r.PathValue("slug")
					db := r.Context().Value("$db").(*gorm.DB)
					form, err := gorm.G[Form](db).Preload("FormFields", func(p gorm.PreloadBuilder) error {
						p.Order("sort_order ASC, id ASC")
						return nil
					}).Where("slug = ?", slug).First(r.Context())
					if err != nil {
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
		}).
		WithLayer("forms.public_post", views.MethodLayer{
			Method: http.MethodPost,
			Handler: func(inner *views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					slug := r.PathValue("slug")
					db := r.Context().Value("$db").(*gorm.DB)
					form, err := gorm.G[Form](db).Preload("FormFields", func(p gorm.PreloadBuilder) error {
						p.Order("sort_order ASC, id ASC")
						return nil
					}).Where("slug = ?", slug).First(r.Context())
					if err != nil {
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
					if len(fieldErrors) != 0 {
						ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
						inner.RenderPage(w, r.WithContext(ctx))
						return
					}
					raw, err := json.Marshal(values)
					if err != nil {
						slog.Error("forms: marshal answers", "error", err)
						fieldErrors["_form"] = err
						ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
						inner.RenderPage(w, r.WithContext(ctx))
						return
					}
					sub := FormSubmission{FormID: form.ID, Answers: datatypes.JSON(raw)}
					if err := gorm.G[FormSubmission](db).Create(r.Context(), &sub); err != nil {
						slog.Error("forms: create submission", "error", err)
						fieldErrors["_form"] = err
						ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
						inner.RenderPage(w, r.WithContext(ctx))
						return
					}
					loc := ThankYouRedirectURL(&form)
					if PublicSubmitSuccessRedirectURL != nil {
						if u := PublicSubmitSuccessRedirectURL(&form); u != "" {
							loc = u
						}
					}
					views.HtmxRedirect(w, r, loc, http.StatusSeeOther)
				})
			},
		})
}
