package forms

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"gorm.io/datatypes"
	gomponents "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

// SubmissionAnswersDisplay renders stored JSON answers using each field's label and type
// (read-only). Expects Detail[FormSubmission] context with $in containing Form (with FormFields)
// and Answers.
type SubmissionAnswersDisplay struct {
	components.Page
}

func (e SubmissionAnswersDisplay) GetKey() string {
	return e.Key
}

func (e SubmissionAnswersDisplay) GetRoles() []string {
	return e.Roles
}

func (e SubmissionAnswersDisplay) Build(ctx context.Context) gomponents.Node {
	in, ok := ctx.Value("$in").(map[string]any)
	if !ok {
		return ghtml.Div(ghtml.Class("text-error"), gomponents.Text("Missing submission data."))
	}

	form, ok := formFromIn(in)
	if !ok {
		return ghtml.Div(ghtml.Class("text-error"), gomponents.Text("Missing form definition."))
	}

	answers, err := parseAnswersMap(in["Answers"])
	if err != nil {
		return ghtml.Div(ghtml.Class("text-error"), gomponents.Text("Invalid answers payload."))
	}

	fields := append([]FormField(nil), form.FormFields...)
	sort.Slice(fields, func(i, j int) bool {
		if fields[i].SortOrder != fields[j].SortOrder {
			return fields[i].SortOrder < fields[j].SortOrder
		}
		return fields[i].ID < fields[j].ID
	})

	var nodes []gomponents.Node
	shown := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		shown[f.Name] = struct{}{}
		raw := answers[f.Name]
		valStr, classes := formatAnswerForField(f, raw)
		nodes = append(nodes, components.Render(&components.LabelInline{
			Page:  components.Page{Key: e.Key + ".label." + f.Name},
			Title: f.Label,
			Children: []components.PageInterface{
				&components.FieldText{
					Page:    components.Page{Key: e.Key + ".val." + f.Name},
					Getter:  getters.GetterStatic(valStr),
					Classes: classes,
				},
			},
		}, ctx))
	}

	var extra []string
	for k := range answers {
		if _, ok := shown[k]; !ok {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	for _, k := range extra {
		v := answers[k]
		nodes = append(nodes, components.Render(&components.LabelInline{
			Page:  components.Page{Key: e.Key + ".extra." + k},
			Title: k + " (unknown field)",
			Children: []components.PageInterface{
				&components.FieldText{
					Page:    components.Page{Key: e.Key + ".extra.val." + k},
					Getter:  getters.GetterStatic(strings.TrimSpace(fmt.Sprint(v))),
					Classes: "font-mono text-sm whitespace-pre-wrap break-all",
				},
			},
		}, ctx))
	}

	return ghtml.Div(ghtml.Class("flex flex-col gap-3"), gomponents.Group(nodes))
}

func formFromIn(in map[string]any) (Form, bool) {
	raw, ok := in["Form"]
	if !ok || raw == nil {
		return Form{}, false
	}
	switch v := raw.(type) {
	case Form:
		return v, true
	case *Form:
		if v == nil {
			return Form{}, false
		}
		return *v, true
	default:
		return Form{}, false
	}
}

func parseAnswersMap(raw any) (map[string]any, error) {
	if raw == nil {
		return map[string]any{}, nil
	}
	if m, ok := raw.(map[string]any); ok {
		return m, nil
	}
	var b []byte
	switch v := raw.(type) {
	case []byte:
		b = v
	case datatypes.JSON:
		b = []byte(v)
	case string:
		b = []byte(v)
	default:
		return nil, fmt.Errorf("unsupported answers type %T", raw)
	}
	if len(b) == 0 {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func formatAnswerForField(f FormField, raw any) (string, string) {
	empty := "—"
	if raw == nil {
		return empty, ""
	}

	switch f.FieldType {
	case FieldTypeNumber:
		s, ok := formatNumberAnswer(raw)
		if !ok {
			return empty, ""
		}
		if strings.TrimSpace(s) == "" {
			return empty, ""
		}
		return s, ""

	case FieldTypeTextarea:
		s := strings.TrimRight(fmt.Sprint(raw), "\r\n")
		if strings.TrimSpace(s) == "" {
			return empty, ""
		}
		return s, "whitespace-pre-wrap break-words"

	case FieldTypeSelect:
		s := strings.TrimSpace(fmt.Sprint(raw))
		if s == "" {
			return empty, ""
		}
		return s, ""

	case FieldTypeEmail:
		s := strings.TrimSpace(fmt.Sprint(raw))
		if s == "" {
			return empty, ""
		}
		return s, ""

	default:
		s := strings.TrimSpace(fmt.Sprint(raw))
		if s == "" {
			return empty, ""
		}
		return s, ""
	}
}

func formatNumberAnswer(raw any) (string, bool) {
	switch n := raw.(type) {
	case float64:
		if math.Trunc(n) == n && !math.IsInf(n, 0) && !math.IsNaN(n) {
			return strconv.FormatInt(int64(n), 10), true
		}
		return strconv.FormatFloat(n, 'f', -1, 64), true
	case json.Number:
		return n.String(), true
	case int:
		return strconv.Itoa(n), true
	case int64:
		return strconv.FormatInt(n, 10), true
	case string:
		n = strings.TrimSpace(n)
		if n == "" {
			return "", false
		}
		return n, true
	default:
		s := strings.TrimSpace(fmt.Sprint(raw))
		if s == "" {
			return "", false
		}
		return s, true
	}
}
