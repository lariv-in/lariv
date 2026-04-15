package components

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/lariv-in/lago/getters"
)

func TestMultiStepFormBuildCarriesPreviousValues(t *testing.T) {
	when := time.Date(2026, time.April, 15, 14, 30, 0, 0, time.UTC)
	ctx := context.WithValue(context.Background(), getters.ContextKeyError, map[string]error{
		"First": fmt.Errorf("bad first"),
		"_form": fmt.Errorf("bad form"),
	})
	form := MultiStepForm{
		Stage:         getters.Static(1),
		Values:        getters.Static(map[string]any{"First": "alpha", "When": when}),
		MultiStageURL: getters.Static("/wizard"),
		Stages: []FormInterface{
			FormComponent[struct{}]{
				Attr: getters.FormBubbling(getters.Static("wizard")),
				ChildrenInput: []PageInterface{
					InputText{Name: "First", Label: "First"},
					InputDate{Name: "When", Label: "When"},
				},
			},
			FormComponent[struct{}]{
				Attr: getters.FormBubbling(getters.Static("wizard")),
				ChildrenInput: []PageInterface{
					InputText{Name: "Second", Label: "Second"},
				},
				ChildrenAction: []PageInterface{
					ButtonSubmit{Label: "Save"},
				},
			},
			FormComponent[struct{}]{
				Attr: getters.FormBubbling(getters.Static("wizard")),
				ChildrenInput: []PageInterface{
					InputText{Name: "Third", Label: "Third"},
				},
			},
		},
	}

	html := renderNode(t, form.Build(ctx))
	if !strings.Contains(html, `action="/wizard"`) {
		t.Fatalf("expected multistage action on rendered form, got %s", html)
	}
	if !strings.Contains(html, `name="$stage"`) || !strings.Contains(html, `value="1"`) {
		t.Fatalf("expected hidden stage field, got %s", html)
	}
	if !strings.Contains(html, `name="$stage_target"`) || !strings.Contains(html, `value="0"`) {
		t.Fatalf("expected ribbon button for previous stage, got %s", html)
	}
	if !strings.Contains(html, `Step 2`) || !strings.Contains(html, `btn-primary`) {
		t.Fatalf("expected active ribbon state for current stage, got %s", html)
	}
	if !strings.Contains(html, `Step 3`) || !strings.Contains(html, `btn-disabled`) {
		t.Fatalf("expected disabled future ribbon stage, got %s", html)
	}
	if !strings.Contains(html, `name="Second"`) {
		t.Fatalf("expected active stage field in html, got %s", html)
	}
	if !strings.Contains(html, `name="First"`) || !strings.Contains(html, `value="alpha"`) {
		t.Fatalf("expected hidden carry field for previous stage, got %s", html)
	}
	if !strings.Contains(html, `name="When"`) || !strings.Contains(html, `value="2026-04-15"`) {
		t.Fatalf("expected typed hidden carry field for date input, got %s", html)
	}
	if !strings.Contains(html, `name="$error.First"`) || !strings.Contains(html, `value="bad first"`) {
		t.Fatalf("expected hidden carry field for field error, got %s", html)
	}
	if !strings.Contains(html, `name="$error._form"`) || !strings.Contains(html, `value="bad form"`) {
		t.Fatalf("expected hidden carry field for form error, got %s", html)
	}
	if got := strings.Count(html, `border-error`); got != 3 {
		t.Fatalf("expected all stages highlighted for global form error, got %d body=%s", got, html)
	}
}

func TestMultiStepFormBuildHighlightsOnlyStagesWithFieldErrors(t *testing.T) {
	ctx := context.WithValue(context.Background(), getters.ContextKeyError, map[string]error{
		"First": fmt.Errorf("bad first"),
		"Third": fmt.Errorf("bad third"),
	})
	form := MultiStepForm{
		Stage: getters.Static(1),
		Stages: []FormInterface{
			FormComponent[struct{}]{
				ChildrenInput: []PageInterface{
					InputText{Name: "First", Label: "First"},
				},
			},
			FormComponent[struct{}]{
				ChildrenInput: []PageInterface{
					InputText{Name: "Second", Label: "Second"},
				},
			},
			FormComponent[struct{}]{
				ChildrenInput: []PageInterface{
					InputText{Name: "Third", Label: "Third"},
				},
			},
		},
	}

	html := renderNode(t, form.Build(ctx))
	if got := strings.Count(html, `border-error`); got != 2 {
		t.Fatalf("expected only error stages highlighted, got %d body=%s", got, html)
	}
	if !strings.Contains(html, `class="btn btn-sm btn-outline border-2 border-error"`) {
		t.Fatalf("expected previous error stage highlighted, got %s", html)
	}
	if !strings.Contains(html, `class="btn btn-sm btn-disabled border-2 border-error"`) {
		t.Fatalf("expected future error stage highlighted, got %s", html)
	}
	if !strings.Contains(html, `class="btn btn-sm btn-primary">Step 2</button>`) {
		t.Fatalf("expected current non-error stage to keep normal styling, got %s", html)
	}
}

func TestMultiStepFormParseFormIncludesCarryForwardValues(t *testing.T) {
	form := MultiStepForm{
		Stages: []FormInterface{
			FormComponent[struct{}]{
				ChildrenInput: []PageInterface{
					InputText{Name: "First"},
					InputCheckbox{Name: "Enabled"},
				},
			},
			FormComponent[struct{}]{
				ChildrenInput: []PageInterface{
					InputNumber[uint]{Name: "Count"},
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"$stage":  {"1"},
		"First":   {"alpha"},
		"Enabled": {"false"},
		"Count":   {"12"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	values, errs, err := form.ParseForm(req)
	if err != nil {
		t.Fatalf("ParseForm returned error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no field errors, got %#v", errs)
	}
	if got, _ := values["First"].(string); got != "alpha" {
		t.Fatalf("expected First alpha, got %#v", values["First"])
	}
	if got, _ := values["Enabled"].(bool); got {
		t.Fatalf("expected Enabled false, got %#v", values["Enabled"])
	}
	if got, _ := values["Count"].(uint); got != 12 {
		t.Fatalf("expected Count 12, got %#v", values["Count"])
	}
}

func TestMultiStepFormParseTargetStageDefaultsToNextStage(t *testing.T) {
	form := MultiStepForm{
		Stages: []FormInterface{
			FormComponent[struct{}]{},
			FormComponent[struct{}]{},
			FormComponent[struct{}]{},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"$stage": {"0"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatalf("ParseForm failed: %v", err)
	}

	if got := form.ParseTargetStage(req, 0); got != 1 {
		t.Fatalf("expected default next stage, got %d", got)
	}
}

func TestParseMultiStepErrorsIncludesCarriedErrors(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"$stage":        {"1"},
		"$error.First":  {"bad first"},
		"$error._form":  {"bad form"},
		"$error.Empty":  {""},
		"$error_target": {"ignore"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatalf("ParseForm failed: %v", err)
	}

	errors := ParseMultiStepErrors(req)
	if got := errors["First"]; got == nil || got.Error() != "bad first" {
		t.Fatalf("expected carried First error, got %#v", got)
	}
	if got := errors["_form"]; got == nil || got.Error() != "bad form" {
		t.Fatalf("expected carried _form error, got %#v", got)
	}
	if _, ok := errors["Empty"]; ok {
		t.Fatalf("expected blank carried errors to be ignored, got %#v", errors)
	}
}
