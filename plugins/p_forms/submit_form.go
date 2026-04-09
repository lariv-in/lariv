package forms

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	gomponents "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

// ContextKeyPublicLoadedForm is the context key for the *Form loaded on the public submit route
// (used with getters.GetterKey, e.g. ContextKeyPublicLoadedForm+".Slug").
const ContextKeyPublicLoadedForm = "forms_public_form"

// PublicSubmitForm renders and parses the public form. It implements components.FormInterface
// so views.View.ParseForm finds it without static FormComponent children.
type PublicSubmitForm struct {
	components.Page
	ActionURL getters.Getter[string]
}

func (e PublicSubmitForm) GetKey() string {
	return e.Key
}

func (e PublicSubmitForm) GetRoles() []string {
	return e.Roles
}

func (e PublicSubmitForm) Build(ctx context.Context) gomponents.Node {
	form, err := getters.Key[*Form](ContextKeyPublicLoadedForm)(ctx)
	if err != nil || form == nil {
		slog.Error("PublicSubmitForm: missing loaded form in context")
		return ghtml.Div(ghtml.Class("text-error"), gomponents.Text("Form unavailable."))
	}

	submitted := ""
	if g, ok := ctx.Value("$get").(map[string]any); ok {
		if v, ok := g["submitted"].(string); ok {
			submitted = v
		}
	}

	action := ""
	if e.ActionURL != nil {
		u, err := e.ActionURL(ctx)
		if err != nil {
			slog.Error("PublicSubmitForm ActionURL failed", "error", err)
		} else {
			action = u
		}
	}

	var header []gomponents.Node
	if submitted == "1" {
		header = append(header, ghtml.Div(ghtml.Class("alert alert-success mb-4"), gomponents.Text("Thank you — your response was submitted.")))
	}

	header = append(header,
		ghtml.Div(ghtml.Class("text-xl font-semibold mb-1"), gomponents.Text(form.Title)),
	)
	if form.Description != "" {
		header = append(header, ghtml.Div(ghtml.Class("text-sm opacity-80 mb-4 whitespace-pre-wrap"), gomponents.Text(form.Description)))
	}

	var fields []gomponents.Node
	for _, f := range form.FormFields {
		fields = append(fields, e.buildField(ctx, f)...)
	}

	return ghtml.Form(
		ghtml.Class("flex flex-col gap-2"),
		ghtml.Method(http.MethodPost),
		gomponents.If(action != "", gomponents.Attr("action", action)),
		gomponents.Group(header),
		gomponents.Group(fields),
		ghtml.Div(ghtml.Class("mt-4"),
			ghtml.Button(ghtml.Class("btn btn-primary"), ghtml.Type("submit"), gomponents.Text("Submit")),
		),
	)
}

func (e PublicSubmitForm) buildField(ctx context.Context, f FormField) []gomponents.Node {
	name := f.Name
	errKey := getters.ContextKeyError + "." + name
	getter := getters.Getter[string](func(c context.Context) (string, error) {
		in := map[string]any{}
		if m, ok := c.Value(getters.ContextKeyIn).(map[string]any); ok {
			in = m
		}
		if v, ok := in[name]; ok {
			if s, ok := v.(string); ok {
				return s, nil
			}
			return fmt.Sprint(v), nil
		}
		return "", nil
	})

	wrap := func(child components.PageInterface) gomponents.Node {
		return components.ContainerError{
			Error:    getters.Key[error](errKey),
			Children: []components.PageInterface{child},
		}.Build(ctx)
	}

	switch f.FieldType {
	case "textarea":
		return []gomponents.Node{wrap(&components.InputTextarea{
			Page:     components.Page{Key: "forms.public." + name},
			Label:    f.Label,
			Name:     name,
			Required: f.Required,
			Rows:     4,
			Getter:   getter,
		})}
	case "email":
		return []gomponents.Node{wrap(&components.InputEmail{
			Page:     components.Page{Key: "forms.public." + name},
			Label:    f.Label,
			Name:     name,
			Required: f.Required,
			Getter:   getter,
		})}
	case "number":
		gn := getters.Getter[int](func(c context.Context) (int, error) {
			in := map[string]any{}
			if m, ok := c.Value(getters.ContextKeyIn).(map[string]any); ok {
				in = m
			}
			if v, ok := in[name]; ok {
				switch n := v.(type) {
				case int:
					return n, nil
				case float64:
					return int(n), nil
				case string:
					i, err := strconv.Atoi(n)
					if err != nil {
						return 0, nil
					}
					return i, nil
				}
			}
			return 0, nil
		})
		return []gomponents.Node{wrap(&components.InputNumber[int]{
			Page:     components.Page{Key: "forms.public." + name},
			Label:    f.Label,
			Name:     name,
			Required: f.Required,
			Getter:   gn,
		})}
	case "select":
		opts := selectOptionsFromField(f)
		choices := getters.Static(opts)
		selGetter := getters.Getter[registry.Pair[string, string]](func(c context.Context) (registry.Pair[string, string], error) {
			in := map[string]any{}
			if m, ok := c.Value(getters.ContextKeyIn).(map[string]any); ok {
				in = m
			}
			if v, ok := in[name]; ok {
				s := fmt.Sprint(v)
				return registry.Pair[string, string]{Key: s, Value: s}, nil
			}
			return registry.Pair[string, string]{}, nil
		})
		return []gomponents.Node{wrap(&components.InputSelect[string]{
			Page:     components.Page{Key: "forms.public." + name},
			Label:    f.Label,
			Name:     name,
			Choices:  choices,
			Getter:   selGetter,
			Required: f.Required,
		})}
	default: // text and unknown
		return []gomponents.Node{wrap(&components.InputText{
			Page:     components.Page{Key: "forms.public." + name},
			Label:    f.Label,
			Name:     name,
			Required: f.Required,
			Getter:   getter,
		})}
	}
}

func selectOptionsFromField(f FormField) []registry.Pair[string, string] {
	var out []registry.Pair[string, string]
	for _, s := range f.SelectOptionStrings() {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		out = append(out, registry.Pair[string, string]{Key: s, Value: s})
	}
	return out
}

func (e PublicSubmitForm) ParseForm(r *http.Request) (map[string]any, map[string]error, error) {
	ctx := r.Context()
	form, err := getters.Key[*Form](ContextKeyPublicLoadedForm)(ctx)
	if err != nil || form == nil {
		return nil, nil, fmt.Errorf("missing form in context")
	}
	if err := r.ParseForm(); err != nil {
		return nil, nil, err
	}
	values := map[string]any{}
	errs := map[string]error{}

	for _, f := range form.FormFields {
		v := r.Form[f.Name]
		raw := ""
		if len(v) > 0 {
			raw = v[0]
		}
		if f.Required && strings.TrimSpace(raw) == "" {
			errs[f.Name] = fmt.Errorf("required")
			continue
		}
		if strings.TrimSpace(raw) == "" && !f.Required {
			values[f.Name] = ""
			continue
		}
		switch f.FieldType {
		case "number":
			n, err := strconv.Atoi(raw)
			if err != nil {
				errs[f.Name] = fmt.Errorf("invalid number")
				continue
			}
			values[f.Name] = n
		case "select":
			ok := false
			for _, opt := range selectOptionsFromField(f) {
				if opt.Key == raw {
					ok = true
					break
				}
			}
			if !ok {
				errs[f.Name] = fmt.Errorf("invalid choice")
				continue
			}
			values[f.Name] = raw
		default:
			values[f.Name] = raw
		}
	}

	return values, errs, nil
}

// AnswersJSON marshals answer map to datatypes.JSON-compatible bytes.
func AnswersJSON(values map[string]any) ([]byte, error) {
	return json.Marshal(values)
}

// ThankYouRedirectURL appends submitted=1 to the public form URL.
func ThankYouRedirectURL(form *Form) string {
	u := "/forms/public/p/" + url.PathEscape(form.Slug) + "/"
	return u + "?" + url.Values{"submitted": {"1"}}.Encode()
}

// PublicSubmitSuccessRedirectURL, if non-nil, is used instead of ThankYouRedirectURL after a successful
// public form POST. Return "" to fall back to the default thank-you URL.
var PublicSubmitSuccessRedirectURL func(form *Form) string
